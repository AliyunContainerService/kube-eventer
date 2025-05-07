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

package influxdb

import (
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/util"
	"net/url"
	"strings"
	"sync"
	"time"

	influxdb_common "github.com/AliyunContainerService/kube-eventer/common/influxdb"
	"github.com/AliyunContainerService/kube-eventer/core"
	metrics_core "github.com/AliyunContainerService/kube-eventer/metrics/core"
	influxdb "github.com/influxdata/influxdb/client"
	kube_api "k8s.io/api/core/v1"
	"k8s.io/klog"
	jsoniter "github.com/json-iterator/go"
)

type influxdbSink struct {
	client influxdb_common.InfluxdbClient
	sync.RWMutex
	c        influxdb_common.InfluxdbConfig
	dbExists bool
}

// eventStruct influxDB field more detial.
type eventStruct struct{
		kind string
		evenType string
		nameSpace string
		podName string
		reason string
		message string
		firstTimestamp string 
		lastTimestamp string 
}

const (
	eventMeasurementName = "log/events"
	// Event special tags
	eventUID = "uid"
	// Value Field name
	valueField = "value"

	// Kubernetes Event const
	kubeKind = "kind"
	evenType = "type"
	nameSpace = "nameSpace"
	podName = "podName"
	eventReason = "reason"
	eventMessage = "message"
	firstTimestamp = "firstTimestamp"
	lastTimestamp = "lastTimestamp"
	
	// Event special tags
	dbNotFoundError = "database not found"

	// Maximum number of influxdb Points to be sent in one batch.
	maxSendBatchSize = 1000
)

func (sink *influxdbSink) resetConnection() {
	klog.Infof("Influxdb connection reset")
	sink.dbExists = false
	sink.client = nil
}

// Generate point value for event
func getEventValue(event *kube_api.Event) (string, error) {
	// TODO: check whether indenting is required.
	bytes, err := json.MarshalIndent(event, "", " ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func eventToPointWithFields(event *kube_api.Event) (*influxdb.Point, error) {
	point := influxdb.Point{
		Measurement: "events",
		Time:        util.GetLastEventTimestamp(event).UTC(),
		Fields: map[string]interface{}{
			"message": event.Message,
		},
		Tags: map[string]string{
			eventUID: string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.Tags[metrics_core.LabelPodId.Key] = string(event.InvolvedObject.UID)
	}
	point.Tags["object_name"] = event.InvolvedObject.Name
	point.Tags["type"] = event.Type
	point.Tags["kind"] = event.InvolvedObject.Kind
	point.Tags["component"] = event.Source.Component
	point.Tags["reason"] = event.Reason
	point.Tags[metrics_core.LabelNamespaceName.Key] = event.Namespace
	point.Tags[metrics_core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}


// selectEventData  Select kubernetes events data from value
func selectEventData(value string) eventStruct{
		var eventData eventStruct
		data := []byte(value)
		eventData.evenType = jsoniter.Get(data,"type").ToString()
		eventData.kind = jsoniter.Get(data,"involvedObject","kind").ToString()
		eventData.nameSpace = jsoniter.Get(data,"involvedObject","namespace").ToString()
		eventData.podName = jsoniter.Get(data,"involvedObject","name").ToString()
		eventData.reason = jsoniter.Get(data,"reason").ToString()
		eventData.message = jsoniter.Get(data,"message").ToString()
		eventData.firstTimestamp = jsoniter.Get(data,"firstTimestamp").ToString()
		eventData.lastTimestamp = jsoniter.Get(data,"lastTimestamp").ToString()
		return eventData
}


// eventToPoint make influxdb point from kubernetes event data
func eventToPoint(event *kube_api.Event) (*influxdb.Point, error) {
	value, err := getEventValue(event)
	if err != nil {
		return nil, err
	}

	eventData := selectEventData(value)
	point := influxdb.Point{
		Measurement: eventMeasurementName,
		Time:        util.GetLastEventTimestamp(event).UTC(),
		Fields: map[string]interface{}{
			valueField: value,
			kubeKind: eventData.kind,
			evenType: eventData.evenType,
			nameSpace: eventData.nameSpace,
			podName: eventData.podName,
			eventReason: eventData.reason,
			eventMessage: eventData.message,
			firstTimestamp: eventData.firstTimestamp,
			lastTimestamp: eventData.lastTimestamp,
		},
		Tags: map[string]string{
			eventUID: string(event.UID),
		},
	}
	if event.InvolvedObject.Kind == "Pod" {
		point.Tags[metrics_core.LabelPodId.Key] = string(event.InvolvedObject.UID)
		point.Tags[metrics_core.LabelPodName.Key] = event.InvolvedObject.Name
	}
	point.Tags[metrics_core.LabelHostname.Key] = event.Source.Host
	return &point, nil
}

func (sink *influxdbSink) ExportEvents(eventBatch *core.EventBatch) {
	sink.Lock()
	defer sink.Unlock()

	dataPoints := make([]influxdb.Point, 0, 10)
	for _, event := range eventBatch.Events {
		var point *influxdb.Point
		var err error
		if sink.c.WithFields {
			point, err = eventToPointWithFields(event)
		} else {
			point, err = eventToPoint(event)
		}
		if err != nil {
			klog.Warningf("Failed to convert event to point: %v", err)
		}

		point.Tags["cluster_name"] = sink.c.ClusterName

		dataPoints = append(dataPoints, *point)
		if len(dataPoints) >= maxSendBatchSize {
			sink.sendData(dataPoints)
			dataPoints = make([]influxdb.Point, 0, 1)
		}
	}
	if len(dataPoints) >= 0 {
		sink.sendData(dataPoints)
	}
}

func (sink *influxdbSink) sendData(dataPoints []influxdb.Point) {
	if err := sink.createDatabase(); err != nil {
		klog.Errorf("Failed to create influxdb: %v", err)
		return
	}
	bp := influxdb.BatchPoints{
		Points:          dataPoints,
		Database:        sink.c.DbName,
		RetentionPolicy: "default",
	}

	start := time.Now()
	if _, err := sink.client.Write(bp); err != nil {
		klog.Errorf("InfluxDB write failed: %v", err)
		if strings.Contains(err.Error(), dbNotFoundError) {
			sink.resetConnection()
		} else if _, _, err := sink.client.Ping(); err != nil {
			klog.Errorf("InfluxDB ping failed: %v", err)
			sink.resetConnection()
		}
	}
	end := time.Now()
	klog.V(4).Infof("Exported %d data to influxDB in %s", len(dataPoints), end.Sub(start))
}

func (sink *influxdbSink) Name() string {
	return "InfluxDB Sink"
}

func (sink *influxdbSink) Stop() {
	// nothing needs to be done.
}

func (sink *influxdbSink) createDatabase() error {
	if sink.client == nil {
		client, err := influxdb_common.NewClient(sink.c)
		if err != nil {
			return err
		}
		sink.client = client
	}

	if sink.dbExists {
		return nil
	}

	q := influxdb.Query{
		Command: fmt.Sprintf(`CREATE DATABASE "%s" WITH NAME "default"`, sink.c.DbName),
	}

	if resp, err := sink.client.Query(q); err != nil {
		// We want to return error only if it is not "already exists" error.
		if !(resp != nil && resp.Err != nil && strings.Contains(resp.Err.Error(), "existing policy")) {
			err := sink.createRetentionPolicy()
			if err != nil {
				return err
			}
		}
	}

	sink.dbExists = true
	klog.Infof("Created database %q on influxDB server at %q", sink.c.DbName, sink.c.Host)
	return nil
}

func (sink *influxdbSink) createRetentionPolicy() error {
	q := influxdb.Query{
		Command: fmt.Sprintf(`CREATE RETENTION POLICY "default" ON "%s" DURATION 0d REPLICATION 1 DEFAULT`, sink.c.DbName),
	}

	if resp, err := sink.client.Query(q); err != nil {
		// We want to return error only if it is not "already exists" error.
		if !(resp != nil && resp.Err != nil && strings.Contains(resp.Err.Error(), "already exists")) {
			return fmt.Errorf("Retention policy creation failed: %v", err)
		}
	}

	klog.Infof("Created database %q on influxDB server at %q", sink.c.DbName, sink.c.Host)
	return nil
}

// Returns a thread-safe implementation of core.EventSink for InfluxDB.
func newSink(c influxdb_common.InfluxdbConfig) core.EventSink {
	client, err := influxdb_common.NewClient(c)
	if err != nil {
		klog.Errorf("issues while creating an InfluxDB sink: %v, will retry on use", err)
	}
	return &influxdbSink{
		client: client, // can be nil
		c:      c,
	}
}

func CreateInfluxdbSink(uri *url.URL) (core.EventSink, error) {
	config, err := influxdb_common.BuildConfig(uri)
	if err != nil {
		return nil, err
	}
	sink := newSink(*config)
	klog.Infof("created influxdb sink with options: host:%s user:%s db:%s", config.Host, config.User, config.DbName)
	return sink, nil
}
