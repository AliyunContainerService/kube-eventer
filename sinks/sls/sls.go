// Copyright 2018 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package sls

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/AliyunContainerService/kube-eventer/core"
	metrics_core "github.com/AliyunContainerService/kube-eventer/metrics/core"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/metadata"
	"github.com/denverdino/aliyungo/sls"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	slsSinkName = "SLSSink"
	eventId     = "eventId"
	podEvent    = "Pod"
	eventLevel  = "level"
	ConfigPath  = "/var/addon/token-config"
)

/*
	Usage:
	--sink=sls:https://sls.aliyuncs.com?logStore=[your_log_store]&project=[your_project_name]
*/
type SLSSink struct {
	Config   *Config
	Project  string
	LogStore string
}

type AKInfo struct {
	AccessKeyId     string `json:"access.key.id"`
	AccessKeySecret string `json:"access.key.secret"`
	SecurityToken   string `json:"security.token"`
	Expiration      string `json:"expiration"`
	Keyring         string `json:"keyring"`
}

// Config can be specific
type Config struct {
	project         string
	logStore        string
	topic           string
	regionId        string
	internal        bool
	accessKeyId     string
	accessKeySecret string
}

func (s *SLSSink) Name() string {
	return slsSinkName
}

func (s *SLSSink) ExportEvents(batch *core.EventBatch) {
	if len(batch.Events) == 0 {
		return
	}
	logs := make([]*sls.Log, 0)
	for _, event := range batch.Events {
		log := &sls.Log{}

		time := uint32(event.LastTimestamp.Unix())

		log.Time = &time

		cts := s.eventToContents(event)

		log.Contents = cts

		logs = append(logs, log)
	}
	request := &sls.PutLogsRequest{
		Project:  s.Project,
		LogStore: s.LogStore,
		LogItems: sls.LogGroup{
			Logs: logs,
		},
	}
	if len(s.Config.topic) > 0 {
		request.LogItems.Topic = &s.Config.topic
	}

	err := s.client().PutLogs(request)
	if err != nil {
		klog.Errorf("failed to put events to sls,because of %s", err.Error())
	}
}

func (s *SLSSink) Stop() {
	//not implement
}

func (s *SLSSink) eventToContents(event *v1.Event) []*sls.Log_Content {
	contents := make([]*sls.Log_Content, 0)
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return nil
	}

	indexKey := eventId
	fullContent := string(bytes)
	contents = append(contents, &sls.Log_Content{
		Key:   &indexKey,
		Value: &fullContent,
	})

	contents = append(contents, &sls.Log_Content{
		Key:   &metrics_core.LabelHostname.Key,
		Value: &event.Source.Host,
	})

	level := eventLevel
	contents = append(contents, &sls.Log_Content{
		Key:   &level,
		Value: &event.Type,
	})

	if event.InvolvedObject.Kind == podEvent {
		podId := string(event.InvolvedObject.UID)
		contents = append(contents, &sls.Log_Content{
			Key:   &metrics_core.LabelPodId.Key,
			Value: &podId,
		})

		contents = append(contents, &sls.Log_Content{
			Key:   &metrics_core.LabelPodName.Key,
			Value: &event.InvolvedObject.Name,
		})
	}

	return contents
}

func (s *SLSSink) client() *sls.Client {
	c, e := newClient(s.Config)
	if e != nil {
		log.Fatalf("can not create sls client because of %s", e.Error())
		return nil
	}
	return c
}

// NewSLSSink returns new SLSSink
func NewSLSSink(uri *url.URL) (*SLSSink, error) {
	s := &SLSSink{}
	c, err := parseConfig(uri)
	if err != nil {
		return nil, err
	}

	s.Project = c.project
	s.LogStore = c.logStore
	s.Config = c
	return s, nil
}

// parseConfig create config from uri
func parseConfig(uri *url.URL) (*Config, error) {
	c := &Config{
		internal: true,
	}

	opts := uri.Query()

	if len(opts["project"]) >= 1 {
		c.project = opts["project"][0]
	} else {
		return nil, errors.New("please provide project name of sls.")
	}

	if len(opts["logStore"]) >= 1 {
		c.logStore = opts["logStore"][0]
	} else {
		return nil, errors.New("please provide logStore name of sls.")
	}

	if len(opts["topic"]) >= 1 {
		c.topic = opts["topic"][0]
	}

	if len(opts["regionId"]) >= 1 {
		c.regionId = opts["regionId"][0]
	} else {
		c.regionId = os.Getenv("RegionId")
	}

	if len(opts["accessKeyId"]) >= 1 {
		c.accessKeyId = opts["accessKeyId"][0]
	} else {
		c.accessKeyId = os.Getenv("AccessKeyId")
	}

	if len(opts["accessKeySecret"]) >= 1 {
		c.accessKeySecret = opts["accessKeySecret"][0]
	} else {
		c.accessKeySecret = os.Getenv("AccessKeySecret")
	}

	if len(opts["internal"]) >= 1 {
		internal, err := strconv.ParseBool(opts["internal"][0])
		if err == nil {
			c.internal = internal
		}
	}
	return c, nil
}

// newClient creates client using AK or metadata
func newClient(c *Config) (*sls.Client, error) {
	m := metadata.NewMetaData(nil)
	region, err := GetRegionFromEnv()
	if err != nil {
		region, err = m.Region()
		if err != nil {
			klog.Errorf("failed to get Region,because of %s", err.Error())
			return nil, err
		}
	}

	var akInfo AKInfo
	if _, err := os.Stat(ConfigPath); err == nil {
		//获取token config json
		encodeTokenCfg, err := ioutil.ReadFile(ConfigPath)
		if err != nil {
			klog.Fatalf("failed to read token config, err: %v", err)
		}
		err = json.Unmarshal(encodeTokenCfg, &akInfo)
		if err != nil {
			klog.Fatalf("error unmarshal token config: %v", err)
		}
		keyring := akInfo.Keyring
		ak, err := Decrypt(akInfo.AccessKeyId, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode ak, err: %v", err)
		}

		sk, err := Decrypt(akInfo.AccessKeySecret, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode sk, err: %v", err)
		}

		token, err := Decrypt(akInfo.SecurityToken, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode token, err: %v", err)
		}
		layout := "2006-01-02T15:04:05Z"
		t, err := time.Parse(layout, akInfo.Expiration)
		if err != nil {
			fmt.Errorf(err.Error())
		}
		if t.Before(time.Now()) {
			klog.Errorf("invalid token which is expired")
		}
		klog.Info("get token by ram role.")
		akInfo.AccessKeyId = string(ak)
		akInfo.AccessKeySecret = string(sk)
		akInfo.SecurityToken = string(token)
	} else {
		roleName, err := m.RoleName()
		if err != nil {
			klog.Errorf("failed to get RoleName,because of %s", err.Error())
			return nil, err
		}

		auth, err := m.RamRoleToken(roleName)
		if err != nil {
			klog.Errorf("failed to get RamRoleToken,because of %s", err.Error())
			return nil, err
		}
		akInfo.AccessKeyId = auth.AccessKeyId
		akInfo.AccessKeySecret = auth.AccessKeySecret
		akInfo.SecurityToken = auth.SecurityToken
	}

	client := sls.NewClientForAssumeRole(common.Region(region), c.internal, akInfo.AccessKeyId, akInfo.AccessKeySecret, akInfo.SecurityToken)
	return client, nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func Decrypt(s string, keyring []byte) ([]byte, error) {
	cdata, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		klog.Errorf("failed to decode base64 string, err: %v", err)
		return nil, err
	}
	block, err := aes.NewCipher(keyring)
	if err != nil {
		klog.Errorf("failed to new cipher, err: %v", err)
		return nil, err
	}
	blockSize := block.BlockSize()

	iv := cdata[:blockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(cdata)-blockSize)

	blockMode.CryptBlocks(origData, cdata[blockSize:])

	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func GetRegionFromEnv() (region string, err error) {
	region = os.Getenv("RegionId")
	if region == "" {
		return "", errors.New("not found region info in env")
	}
	return region, nil
}
