// Copyright 2015 Google Inc. All Rights Reserved.
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

package rabbitmq

import (
	"encoding/json"
	"github.com/AliyunContainerService/kube-eventer/util"
	"net/url"
	"sync"
	"time"

	"k8s.io/klog"

	rabbitmq_common "github.com/AliyunContainerService/kube-eventer/common/rabbitmq"
	event_core "github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/metrics/core"
	kube_api "k8s.io/api/core/v1"
)

type RabbitmqSinkPoint struct {
	EventValue     interface{}
	EventTimestamp time.Time
	EventTags      map[string]string
}

type rabbitmqSink struct {
	rabbitmq_common.AmqpClient
	sync.RWMutex
}

func getEventValue(event *kube_api.Event) (string, error) {
	// TODO: check whether indenting is required.
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func eventToPoint(event *kube_api.Event) (*RabbitmqSinkPoint, error) {
	value, err := getEventValue(event)
	if err != nil {
		return nil, err
	}
	point := RabbitmqSinkPoint{
		EventTimestamp: util.GetLastEventTimestamp(event).UTC(),
		EventValue:     value,
		EventTags: map[string]string{
			"eventID": string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.EventTags[core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.EventTags[core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.EventTags[core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}

func (sink *rabbitmqSink) ExportEvents(eventBatch *event_core.EventBatch) {
	sink.Lock()
	defer sink.Unlock()

	for _, event := range eventBatch.Events {
		point, err := eventToPoint(event)
		if err != nil {
			klog.Warningf("Failed to convert event to point: %v", err)
		}

		err = sink.ProduceAmqpMessage(*point)
		if err != nil {
			klog.Errorf("Failed to produce event message: %s", err)
		}
	}
}

func NewRabbitmqSink(uri *url.URL) (event_core.EventSink, error) {
	client, err := rabbitmq_common.NewAmqpClient(uri, rabbitmq_common.EventsTopic)
	if err != nil {
		return nil, err
	}

	return &rabbitmqSink{
		AmqpClient: client,
	}, nil
}
