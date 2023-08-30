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

package logsink

import (
	"bytes"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/util"

	"github.com/AliyunContainerService/kube-eventer/core"
	"k8s.io/klog"
)

type LogSink struct {
}

func (this *LogSink) Name() string {
	return "LogSink"
}

func (this *LogSink) Stop() {
	// Do nothing.
}

func batchToString(batch *core.EventBatch) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("EventBatch     Timestamp: %s\n", batch.Timestamp))
	for _, event := range batch.Events {
		buffer.WriteString(fmt.Sprintf("%++v   %s (cnt:%d): %s\n", event, util.GetLastEventTimestamp(event), event.Count, event.Message))
	}
	return buffer.String()
}

func (this *LogSink) ExportEvents(batch *core.EventBatch) {
	klog.Info(batchToString(batch))
}

func CreateLogSink() (*LogSink, error) {
	return &LogSink{}, nil
}
