// Copyright 2018 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/AliyunContainerService/kube-eventer/core"
	metrics_core "github.com/AliyunContainerService/kube-eventer/metrics/core"
	"github.com/AliyunContainerService/kube-eventer/sinks/utils"
	"github.com/AliyunContainerService/kube-eventer/util"
	sls "github.com/aliyun/aliyun-log-go-sdk"
	sls_producer "github.com/aliyun/aliyun-log-go-sdk/producer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	slsSinkName        = "SLSSink"
	eventId            = "eventId"
	podEvent           = "Pod"
	eventLevel         = "level"
	SLSDefaultEndpoint = "log.aliyuncs.com"
	SLSUserAgent       = "ack-kube-eventer"

	MaxLogGroupInBytes = 5 * 1024 * 1024 // 5MB
)

/*
 * Usage:
 * --sink=sls:https://sls.aliyuncs.com?logStore=[your_log_store]&project=[your_project_name]&label=<key,value>
 */
type SLSSink struct {
	Config   *Config
	Project  string
	LogStore string

	// current sls producer
	Producer *sls_producer.Producer

	// current ak_info with expiration time
	AkInfo *utils.AKInfo
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
	label           map[string]string
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

		time := uint32(util.GetLastEventTimestamp(event).Unix())

		log.Time = &time

		cts := eventToContents(event, s.Config.label)

		log.Contents = cts

		logs = append(logs, log)
	}

	if err := s.SendLogs(logs); err != nil {
		klog.Errorf("failed to export events to sls, because of %v", err)
		return
	}
}

// SendLogs send logs to sls.
func (s *SLSSink) SendLogs(logs []*sls.Log) (err error) {
	nextSendIndex := 0 // the start index of next send logs
	size := 0

	// send logs in batches
	for i, log := range logs {
		if log.Size() > MaxLogGroupInBytes {
			return fmt.Errorf("send logs [logs size: %v], %w", log.Size(), errors.New("the size of log too large"))
		}
		size += log.Size()
		if size > MaxLogGroupInBytes {
			err = s.getProducer().SendLogListWithCallBack(s.Project, s.LogStore, s.Config.topic, "", logs[nextSendIndex:i], callback{})
			if err != nil {
				return fmt.Errorf("send logs in batches [logs size: %v], err: %w", size, err)
			}
			nextSendIndex = i
			size = log.Size()
		}
	}

	// send remaining logs
	err = s.getProducer().SendLogListWithCallBack(s.Project, s.LogStore, s.Config.topic, "", logs[nextSendIndex:], callback{})
	if err != nil {
		return fmt.Errorf("send remaining logs [logs size: %v], err: %w", size, err)
	}
	return nil
}

func (s *SLSSink) Stop() {
	// safe close producer: close after all data is sent
	s.getProducer().SafeClose()
}

// get a sls producer.
// if akInfo expiration, recreate a new producer.
func (s *SLSSink) getProducer() *sls_producer.Producer {
	if s.Producer == nil {
		klog.Error("get producer, err: %w", errors.New("producer is nil"))
		return nil
	}

	if s.AkInfo.IsExpired() {
		// if akInfo expiration, recreate a new producer.
		klog.Infof("akInfo is expiration, start to recreate a new producer.")
		// 1. stop the old producer
		s.Producer.SafeClose()
		// 2. create a new producer
		newProducer, newAkInfo, err := newProducer(s.Config)
		if err != nil {
			klog.Errorf("failed to recreate new producer, because of %v", err)
			return nil
		}
		s.Producer = newProducer
		s.AkInfo = newAkInfo
		// 3. start the new producer
		if s.Producer != nil {
			s.Producer.Start()
		}
		klog.Infof("recreate new producer, when akInfo expiration")
	}
	return s.Producer
}

func eventToContents(event *v1.Event, labels map[string]string) []*sls.LogContent {
	contents := make([]*sls.LogContent, 0)
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return nil
	}

	indexKey := eventId
	fullContent := string(bytes)
	contents = append(contents, &sls.LogContent{
		Key:   &indexKey,
		Value: &fullContent,
	})

	contents = append(contents, &sls.LogContent{
		Key:   &metrics_core.LabelHostname.Key,
		Value: &event.Source.Host,
	})

	level := eventLevel
	contents = append(contents, &sls.LogContent{
		Key:   &level,
		Value: &event.Type,
	})

	if event.InvolvedObject.Kind == podEvent {
		podId := string(event.InvolvedObject.UID)
		contents = append(contents, &sls.LogContent{
			Key:   &metrics_core.LabelPodId.Key,
			Value: &podId,
		})

		contents = append(contents, &sls.LogContent{
			Key:   &metrics_core.LabelPodName.Key,
			Value: &event.InvolvedObject.Name,
		})
	}

	for key, value := range labels {
		// deep copy
		newKey := key
		newValue := value
		contents = append(contents, &sls.LogContent{
			Key:   &newKey,
			Value: &newValue,
		})
	}

	return contents
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

	producer, akInfo, err := newProducer(c)
	if err != nil {
		return nil, err
	}
	s.Producer = producer
	s.AkInfo = akInfo
	if s.Producer != nil {
		s.Producer.Start()
	}
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

	if len(opts["label"]) >= 1 {
		labelsStrs := opts["label"]
		c.label = parseLabels(labelsStrs)
	}

	return c, nil
}

func parseLabels(labelsStrs []string) map[string]string {
	labels := make(map[string]string)
	for _, kv := range labelsStrs {
		kvItems := strings.Split(kv, ",")
		if len(kvItems) == 2 {
			labels[kvItems[0]] = kvItems[1]
		} else {
			klog.Errorf("parse sls labels error. labelsStr: %v, kv format error: %v", labelsStrs, kv)
		}
	}
	return labels
}

// newProducer create producer with config and new akInfo.
func newProducer(c *Config) (*sls_producer.Producer, *utils.AKInfo, error) {
	// get region from env
	region, parseEnvErr := utils.GetRegionFromEnv()
	if parseEnvErr != nil {
		if c.regionId != "" {
			// region from client
			region = c.regionId
		} else {
			// region from meta data
			regionInMeta, err := utils.ParseRegionFromMeta()
			if err != nil {
				klog.Errorf("failed to get Region, because of %v", err)
				return nil, nil, err
			}
			region = regionInMeta
		}
	}

	// get ak info
	akInfo, err := utils.ParseAKInfoFromConfigPath()
	if err != nil {
		if len(c.accessKeyId) > 0 && len(c.accessKeySecret) > 0 {
			akInfo.AccessKeyId = c.accessKeyId
			akInfo.AccessKeySecret = c.accessKeySecret
			akInfo.SecurityToken = ""
		} else {
			akInfoInMeta, err := utils.ParseAKInfoFromMeta()
			if err != nil {
				klog.Errorf("failed to get RamRoleToken,because of %v", err)
				return nil, nil, err
			}
			akInfo = akInfoInMeta
		}
	}

	// construct sls producer config
	cfg := sls_producer.GetDefaultProducerConfig()
	cfg.Endpoint = getSLSEndpoint(region, c.internal)
	cfg.Region = region
	cfg.UserAgent = SLSUserAgent
	cfg.CredentialsProvider = sls.NewStaticCredentialsProvider(akInfo.AccessKeyId, akInfo.AccessKeySecret, akInfo.SecurityToken)
	cfg.AuthVersion = sls.AuthV4
	producer := sls_producer.InitProducer(cfg)
	return producer, akInfo, nil
}

// refer doc: https://help.aliyun.com/zh/sls/developer-reference/endpoints
func getSLSEndpoint(region string, internal bool) string {
	finalEndpoint := SLSDefaultEndpoint
	endpointFromEnv := os.Getenv("SLS_ENDPOINT")
	if endpointFromEnv != "" {
		finalEndpoint = endpointFromEnv
	}

	if internal {
		region = fmt.Sprintf("%s-intranet", region)
		finalEndpoint = fmt.Sprintf("%s.%s", region, SLSDefaultEndpoint)
	}
	klog.V(6).Infof("sls endpoint, %v", finalEndpoint)
	return finalEndpoint
}

// callback, use it to implement the sls_producer.Callback interface
// to obtain the result of each send,
// because the producer sends requests to the server asynchronously.
type callback struct {
}

func (c callback) Success(result *sls_producer.Result) {
	klog.V(6).Infof("Successfully used Producer to send log list")
}

func (c callback) Fail(result *sls_producer.Result) {
	if result == nil {
		klog.Error("producer failed to send requests, but result is nil")
		return
	}
	klog.Errorf("Failed to send log list using Producer. "+
		"ErrorCode: %v, ErrorMessage: %v, RequestID: %v, Timestamp: %v",
		result.GetErrorCode(),
		result.GetErrorMessage(),
		result.GetRequestId(),
		result.GetTimeStampMs())
}
