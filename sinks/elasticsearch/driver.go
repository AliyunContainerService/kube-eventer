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

package elasticsearch

import (
	"net/url"
	"sync"
	"time"

	esCommon "github.com/AliyunContainerService/kube-eventer/common/elasticsearch"
	event_core "github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/metrics/core"
	"github.com/prometheus/client_golang/prometheus"
	kube_api "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	typeName = "events"
)

// SaveDataFunc is a pluggable function to enforce limits on the object
type SaveDataFunc func(date time.Time, namespace string, sinkData []interface{}) error

type elasticSearchSink struct {
	esSvc     esCommon.ElasticSearchService
	saveData  SaveDataFunc
	flushData func() error
	sync.RWMutex
	errorRate prometheus.Gauge
}

type EsSinkPoint struct {
	Count                    interface{}
	Metadata                 interface{}
	InvolvedObject           interface{}
	Source                   interface{}
	FirstOccurrenceTimestamp time.Time
	LastOccurrenceTimestamp  time.Time
	Message                  string
	Reason                   string
	Type                     string
	EventTags                map[string]string
}

func eventToPoint(event *kube_api.Event, clusterName string) (*EsSinkPoint, error) {
	var (
		lastOccurrenceTimestamp  = event.LastTimestamp.Time.UTC()
		firstOccurrenceTimestamp = event.FirstTimestamp.Time.UTC()
	)

	// Part of k8s resources FirstOccurrenceTimestamp/LastOccurrenceTimestamp is nil
	if event.LastTimestamp.UTC().IsZero() {
		lastOccurrenceTimestamp = event.CreationTimestamp.Time.UTC()
	}

	if event.FirstTimestamp.UTC().IsZero() {
		firstOccurrenceTimestamp = event.CreationTimestamp.Time.UTC()
	}

	point := EsSinkPoint{
		FirstOccurrenceTimestamp: firstOccurrenceTimestamp,
		LastOccurrenceTimestamp:  lastOccurrenceTimestamp,
		Message:                  event.Message,
		Reason:                   event.Reason,
		Type:                     event.Type,
		Count:                    event.Count,
		Metadata:                 event.ObjectMeta,
		InvolvedObject:           event.InvolvedObject,
		Source:                   event.Source,
		EventTags: map[string]string{
			"eventID":      string(event.UID),
			"cluster_name": clusterName,
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.EventTags[core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.EventTags[core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.EventTags[core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}

func (sink *elasticSearchSink) ExportEvents(eventBatch *event_core.EventBatch) {
	var namespace string
	sink.Lock()
	defer sink.Unlock()
	for _, event := range eventBatch.Events {
		point, err := eventToPoint(event, sink.esSvc.ClusterName)
		if err != nil {
			klog.Warningf("Failed to convert event to point: %v", err)
		}
		if sink.esSvc.UseNamespace {
			namespace = event.Namespace
		}
		err = sink.saveData(point.LastOccurrenceTimestamp, namespace, []interface{}{*point})
		if err != nil {
			klog.Warningf("Failed to export data to ElasticSearch sink: %v", err)
		}
	}

	err := sink.flushData()
	if err != nil {
		klog.Warningf("Failed to flushing data to ElasticSearch sink: %v", err)
	}
	if sink.errorRate != nil {
		sink.errorRate.Set(float64(sink.esSvc.ErrorStats()))
	}
}

func (sink *elasticSearchSink) Name() string {
	return "ElasticSearch Sink"
}

func (sink *elasticSearchSink) Stop() {
	// nothing needs to be done.
}

func NewElasticSearchSink(uri *url.URL) (event_core.EventSink, error) {
	var esSink elasticSearchSink
	esSvc, err := esCommon.CreateElasticSearchService(uri)
	if err != nil {
		klog.Warning("Failed to config ElasticSearch")
		return nil, err
	}

	esSink.esSvc = *esSvc
	esSink.saveData = func(date time.Time, namespace string, sinkData []interface{}) error {
		return esSvc.SaveData(date, typeName, namespace, sinkData)
	}
	esSink.flushData = func() error {
		return esSvc.FlushData()
	}

	esSink.errorRate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "eventer",
			Subsystem: "elasticsearch",
			Name:      "errors",
			Help:      "Bulk processing errors.",
		})
	prometheus.MustRegister(esSink.errorRate)

	klog.V(2).Info("ElasticSearch sink setup successfully")
	return &esSink, nil
}
