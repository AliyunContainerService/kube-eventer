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

// EventProcessor is used to process every event, eg filter, transfer...
type EventProcessor interface {
	Process(*EventBatch) *EventBatch
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
