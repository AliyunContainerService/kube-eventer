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
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"strconv"

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

// Config can be specific
type Config struct {
	accessKeyId     string
	accessKeySecret string
	token           string
	project         string
	logStore        string
	topic           string
	internal        bool
	regionId        string
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

	if len(opts["accessKeyId"]) >= 1 {
		c.accessKeyId = opts["accessKeyId"][0]
	} else {
		c.accessKeyId = os.Getenv("AccessKeyId")
	}

	if len(opts["internal"]) >= 1 {
		internal, err := strconv.ParseBool(opts["internal"][0])
		if err == nil {
			c.internal = internal
		}
	}

	if len(opts["accessKeySecret"]) >= 1 {
		c.accessKeySecret = opts["accessKeySecret"][0]
	} else {
		c.accessKeySecret = os.Getenv("AccessKeySecret")
	}

	if len(opts["regionId"]) >= 1 {
		c.regionId = opts["regionId"][0]
	} else {
		c.regionId = os.Getenv("RegionId")
	}

	return c, nil
}

// newClient creates client using AK or metadata
func newClient(c *Config) (*sls.Client, error) {
	if c.regionId == "" || c.accessKeyId == "" || c.accessKeySecret == "" {
		klog.Info("accessKeyId,accessKeySecret or regionId is empty and use metadata instead.")

		m := metadata.NewMetaData(nil)
		regionId, err := m.Region()
		if err != nil {
			klog.Errorf("failed to get Region,because of %s", err.Error())
			return nil, err
		}

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

		client := sls.NewClientForAssumeRole(common.Region(regionId), c.internal, auth.AccessKeyId, auth.AccessKeySecret, auth.SecurityToken)
		return client, nil
	}
	return sls.NewClient(common.Region(c.regionId), c.internal, c.accessKeyId, c.accessKeySecret), nil
}
