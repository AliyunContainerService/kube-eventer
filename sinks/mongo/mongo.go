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

package mongo

import (
	"context"
	"github.com/AliyunContainerService/kube-eventer/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	kube_api "k8s.io/api/core/v1"
	"k8s.io/klog"
	"net/url"
	"sync"
	"time"
)

type mongoSink struct {
	client  *mongo.Client
	closeDB func()
	sync.RWMutex
}

type mongoSinkPoint struct {
	Count                    int32     `bson:"count,omitempty"`
	Namespace                string    `bson:"namespace,omitempty"`
	Kind                     string    `bson:"kind,omitempty"`
	Name                     string    `bson:"name,omitempty"`
	Type                     string    `bson:"type,omitempty"`
	Reason                   string    `bson:"reason,omitempty"`
	Message                  string    `bson:"message,omitempty"`
	EventID                  string    `bson:"event_id,omitempty"`
	FirstOccurrenceTimestamp time.Time `bson:"first_occurrence_time,omitempty"`
	LastOccurrenceTimestamp  time.Time `bson:"last_occurrence_time,omitempty"`
}

func (m *mongoSink) Name() string {
	return "Mongo Sink"
}

func (m *mongoSink) saveData(sinkData *mongoSinkPoint) error {
	eventCollection := m.client.Database("k8s").Collection("event")

	ctx := context.TODO()
	_, err := eventCollection.InsertOne(ctx, sinkData)
	if err != nil {
		return err
	}
	return nil
}

func eventToPoint(event *kube_api.Event) (*mongoSinkPoint, error) {
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

	point := mongoSinkPoint{
		Count:                    event.Count,
		Name:                     event.InvolvedObject.Name,
		Namespace:                event.InvolvedObject.Namespace,
		EventID:                  string(event.UID),
		Type:                     event.Type,
		Reason:                   event.Reason,
		Message:                  event.Message,
		Kind:                     event.InvolvedObject.Kind,
		FirstOccurrenceTimestamp: firstOccurrenceTimestamp,
		LastOccurrenceTimestamp:  lastOccurrenceTimestamp,
	}
	return &point, nil
}

func (m *mongoSink) ExportEvents(eventBatch *core.EventBatch) {
	m.Lock()
	defer m.Unlock()

	for _, event := range eventBatch.Events {
		point, err := eventToPoint(event)
		if err != nil {
			klog.Warningf("Failed to convert event to point: %v", err)
			continue
		}

		err = m.saveData(point)
		if err != nil {
			klog.Warningf("Failed to export data to Mongo sink: %v", err)
		}
	}

}

func (m *mongoSink) Stop() {
	m.closeDB()
}

func CreateMongoSink(uri *url.URL) (core.EventSink, error) {
	var sink mongoSink

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri.RawQuery))
	if err != nil {
		return nil, err
	}
	sink.client = client
	sink.closeDB = func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}

	klog.V(3).Info("Mongo Sink setup successfully")
	return &sink, nil
}
