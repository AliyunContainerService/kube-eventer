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

package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"kube-eventer/core"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	SLACK_SINK            = "SlackSink"
	WARNING           int = 2
	DEFAULT_MSG_TYPE      = "text"
	CONTENT_TYPE_JSON     = "application/json"
	LABE_TEMPLATE         = "%s\n"
)

var (
	MSG_TEMPLATE = "Level:%s \nKind:%s \nNamespace:%s \nName:%s \nReason:%s \nTimestamp:%s \nMessage:%s"

	MSG_TEMPLATE_ARR = [][]string{
		{"Level"},
		{"Kind"},
		{"Namespace"},
		{"Name"},
		{"Reason"},
		{"Timestamp"},
		{"Message"},
	}
)

type SlackMsg struct {
	MsgType string    `json:"msgtype"`
	Text    SlackText `json:"text"`
}

type SlackText struct {
	Content string `json:"content"`
}

type SlackSink struct {
	Color      string
	Icon       string
	MsgType    string
	Token      string
	Username   string
	Level      int
	Labels     []string
	Namespaces []string
	Kinds      []string
}

func (s *SlackSink) Name() string {
	return SLACK_SINK
}

func (s *SlackSink) Face(icon string) {
	s.Icon = icon
}

func (s *SlackSink) Stop() {
	//do nothing
}

func (s *SlackSink) ExportEvents(batch *core.EventBatch) {
	for _, event := range batch.Events {
		if s.isEventLevelDangerous(event.Type) {
			s.Notify(event)
			// add threshold
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (s *SlackSink) isEventLevelDangerous(level string) bool {
	score := getLevel(level)
	if score >= s.Level {
		return true
	}
	return false
}

func (s *SlackSink) Notify(event *v1.Event) {
	if s.Namespaces != nil {
		skip := true
		for _, namespace := range s.Namespaces {
			if namespace == event.Namespace {
				skip = false
				break
			}
		}
		if skip {
			return
		}
	}

	if s.Kinds != nil {
		skip := true
		for _, kind := range s.Kinds {
			if kind == event.InvolvedObject.Kind {
				skip = false
				break
			}
		}
		if skip {
			return
		}
	}

	msg := createMsgFromEvent(s, event)
	if msg == nil {
		klog.Warningf("failed to create msg from event,because of %v", event)
		return
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		klog.Warningf("failed to marshal msg %v", msg)
		return
	}

	b := bytes.NewBuffer(msgBytes)

	resp, err := http.Post(fmt.Sprintf("https://hooks.slack.com/services/%s", s.Token), CONTENT_TYPE_JSON, b)
	if err != nil {
		klog.Errorf("failed to send msg to slack hooks. error: %s", err.Error())
		return
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		klog.Errorf("failed to send msg to slack hooks, because the response code is %d", resp.StatusCode)
		return
	}

}

func getLevel(level string) int {
	score := 0
	switch level {
	case v1.EventTypeWarning:
		score += 2
	case v1.EventTypeNormal:
		score += 1
	default:
		//score will remain 0
	}
	return score
}

func createMsgFromEvent(s *SlackSink, event *v1.Event) *SlackMsg {
	msg := &SlackMsg{}
	msg.MsgType = DEFAULT_MSG_TYPE

	//默认按文本模式推送
	template := MSG_TEMPLATE
	if len(s.Labels) > 0 {
		for _, label := range s.Labels {
			template = fmt.Sprintf(LABE_TEMPLATE, label) + template
		}
	}

	msg.Text = SlackText{
		Content: fmt.Sprintf(template, event.Type, event.InvolvedObject.Kind, event.Namespace, event.Name, event.Reason, event.LastTimestamp.String(), event.Message),
	}
	return msg
}

func NewSlackSink(url *url.URL) (*SlackSink, error) {
	s := &SlackSink{
		Level: WARNING,
	}

	opts := url.Query()

	if len(opts["token"]) >= 1 {
		s.Token = opts["token"][0]
	} else {
		return nil, fmt.Errorf("you must provide slack bot token")
	}

	if len(opts["color"]) >= 1 {
		s.Color = opts["color"][0]
	}

	if len(opts["icon"]) >= 1 {
		s.Color = opts["icon"][0]
	}

	if len(opts["username"]) >= 1 {
		s.Username = opts["username"][0]
	}

	if len(opts["level"]) >= 1 {
		s.Level = getLevel(opts["level"][0])
	}

	if len(opts["label"]) >= 1 {
		s.Labels = opts["label"]
	}

	s.MsgType = DEFAULT_MSG_TYPE

	s.Namespaces = getValues(opts["namespaces"])
	// kinds:https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#lists-and-simple-kinds
	// such as node,pod,component and so on
	s.Kinds = getValues(opts["kinds"])

	return s, nil
}

func getValues(o []string) []string {
	if len(o) >= 1 {
		if len(o[0]) == 0 {
			return nil
		}
		return strings.Split(o[0], ",")
	}
	return nil
}
