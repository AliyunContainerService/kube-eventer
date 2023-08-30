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

package mysql

import (
	"encoding/json"
	mysql_common "github.com/AliyunContainerService/kube-eventer/common/mysql"
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/util"
	kube_api "k8s.io/api/core/v1"
	"k8s.io/klog"
	"net/url"
	"sync"
)

// SaveDataFunc is a pluggable function to enforce limits on the object
type SaveDataFunc func(sinkData []interface{}) error

type mysqlSink struct {
	mysqlSvc  *mysql_common.MysqlService
	saveData  SaveDataFunc
	flushData func() error
	closeDB   func() error
	sync.RWMutex
	uri *url.URL
}

const (
	// Maximum number of mysql Points to be sent in one batch.
	maxSendBatchSize = 1
)

func (sink *mysqlSink) createDatabase() error {

	if sink.mysqlSvc == nil {
		mysqlSvc, err := mysql_common.NewMysqlClient(sink.uri)
		if err != nil {
			return err
		}
		sink.mysqlSvc = mysqlSvc
	}

	return nil
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

func eventToPoint(event *kube_api.Event) (*mysql_common.MysqlKubeEventPoint, error) {

	value, err := getEventValue(event)
	if err != nil {
		return nil, err
	}
	klog.V(9).Infof(value)

	point := mysql_common.MysqlKubeEventPoint{
		Name:                     event.InvolvedObject.Name,
		Namespace:                event.InvolvedObject.Namespace,
		EventID:                  string(event.UID),
		Type:                     event.Type,
		Reason:                   event.Reason,
		Message:                  event.Message,
		Kind:                     event.InvolvedObject.Kind,
		FirstOccurrenceTimestamp: event.FirstTimestamp.Time.String(),
		LastOccurrenceTimestamp:  util.GetLastEventTimestamp(event).String(),
	}

	return &point, nil

}

func (sink *mysqlSink) ExportEvents(eventBatch *core.EventBatch) {

	sink.Lock()
	defer sink.Unlock()

	dataPoints := make([]mysql_common.MysqlKubeEventPoint, 0, 10)
	for _, event := range eventBatch.Events {

		point, err := eventToPoint(event)
		if err != nil {
			klog.Warningf("Failed to convert event to point: %v", err)
			klog.Warningf("Skip this event")
			continue
		}

		dataPoints = append(dataPoints, *point)
		if len(dataPoints) >= maxSendBatchSize {
			err = sink.saveData([]interface{}{*point})
			if err != nil {
				klog.Warningf("Failed to export data to Mysql sink: %v", err)
			}
			dataPoints = make([]mysql_common.MysqlKubeEventPoint, 0, 1)
		}

	}
	klog.V(1).Infof("sinking %v events to mysql success.", len(eventBatch.Events))
}

func (sink *mysqlSink) Name() string {
	return "MySQL Sink"
}

func (sink *mysqlSink) Stop() {
	defer sink.closeDB()
}

// Returns a thread-safe implementation of core.EventSink for InfluxDB.
func CreateMysqlSink(uri *url.URL) (core.EventSink, error) {

	var mySink mysqlSink

	mysqlSvc, err := mysql_common.NewMysqlClient(uri)
	if err != nil {
		return nil, err
	}

	mySink.mysqlSvc = mysqlSvc
	mySink.saveData = func(sinkData []interface{}) error {
		return mysqlSvc.SaveData(sinkData)
	}
	mySink.flushData = func() error {
		return mysqlSvc.FlushData()
	}
	mySink.closeDB = func() error {
		return mysqlSvc.CloseDB()
	}
	mySink.uri = uri

	klog.V(3).Info("Mysql Sink setup successfully")
	return &mySink, nil
}
