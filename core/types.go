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

package core

import (
	"time"

	kube_api "k8s.io/api/core/v1"
)




//DistinctSameResourceEvent : Distinct Events base on involvedObject.Name & event.Reason
func (eventBatch *EventBatch) DistinctSameResourceEvent() {
	tempMap := make(map[string]bool)
	var finalEvents []*kube_api.Event
	for _, event := range eventBatch.Events {
		involvedObject := event.InvolvedObject
		if &involvedObject == nil {
			continue
		}
		resourceName := involvedObject.Name
		reason := event.Reason
		msg:=event.Message
		key := resourceName + reason + msg
		if _, contain := tempMap[key]; !contain {
			// fmt.Printf("key: %s \n", key)
			tempMap[key] = true
			finalEvents = append(finalEvents, event)
		}
	}

	if len(finalEvents) > 0 {
		eventBatch.Events = finalEvents
	}
}


type EventBatch struct {
	// When this batch was created.
	Timestamp time.Time
	// List of events included in the batch.
	Events []*kube_api.Event
}

// A place from where the events should be scraped.
type EventSource interface {
	// This is a mutable method. Each call to this method clears the internal buffer so that
	// each event can be obtained only once.
	GetNewEvents() *EventBatch
}

type EventSink interface {
	Name() string

	// Exports data to the external storage. The function should be synchronous/blocking and finish only
	// after the given EventBatch was written. This will allow sink manager to push data only to these
	// sinks that finished writing the previous data.
	ExportEvents(*EventBatch)
	// Stops the sink at earliest convenience.
	Stop()
}
