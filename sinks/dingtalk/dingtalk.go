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

package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/util"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AliyunContainerService/kube-eventer/core"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	DINGTALK_SINK         = "DingTalkSink"
	WARNING           int = 2
	NORMAL            int = 1
	DEFAULT_MSG_TYPE      = "text"
	CONTENT_TYPE_JSON     = "application/json"
	LABEL_TEMPLATE        = "%s\n"
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

/*
*
dingtalk msg struct
*/
type DingTalkMsg struct {
	MsgType  string           `json:"msgtype"`
	Text     DingTalkText     `json:"text"`
	Markdown DingTalkMarkdown `json:"markdown"`
}

type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DingTalkText struct {
	Content string `json:"content"`
}

/*
*
dingtalk sink usage
--sink:dingtalk:https://oapi.dingtalk.com/robot/send?access_token=[access_token]&level=Warning&label=[label]

level: Normal or Warning. The event level greater than global level will emit.
label: some thing unique when you want to distinguish different k8s clusters.
*/
type DingTalkSink struct {
	Endpoint   string
	Namespaces []string
	Kinds      []string
	Token      string
	Level      int
	Labels     []string
	MsgType    string
	ClusterID  string
	Secret     string
	Region     string
}

func (d *DingTalkSink) Name() string {
	return DINGTALK_SINK
}

func (d *DingTalkSink) Stop() {
	//do nothing
}

func (d *DingTalkSink) ExportEvents(batch *core.EventBatch) {
	for _, event := range batch.Events {
		if d.isEventLevelDangerous(event.Type) {
			d.Ding(event)
			// add threshold
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (d *DingTalkSink) isEventLevelDangerous(level string) bool {
	score := getLevel(level)
	if score >= d.Level {
		return true
	}
	return false
}

func (d *DingTalkSink) Ding(event *v1.Event) {
	value := url.Values{}

	if d.Namespaces != nil {
		skip := true
		for _, namespace := range d.Namespaces {
			if namespace == event.Namespace {
				skip = false
				break
			}
		}
		if skip {
			return
		}
	}

	if d.Kinds != nil {
		skip := true
		for _, kind := range d.Kinds {
			if kind == event.InvolvedObject.Kind {
				skip = false
				break
			}
		}
		if skip {
			return
		}
	}

	msg := createMsgFromEvent(d, event)
	if msg == nil {
		klog.Warningf("failed to create msg from event,because of %v", event)
		return
	}

	msg_bytes, err := json.Marshal(msg)
	if err != nil {
		klog.Warningf("failed to marshal msg %v", msg)
		return
	}

	value.Set("access_token", d.Token)
	if d.Secret != "" {
		t := time.Now().UnixNano() / 1e6
		value.Set("timestamp", fmt.Sprintf("%d", t))
		value.Set("sign", sign(t, d.Secret))
	}

	b := bytes.NewBuffer(msg_bytes)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s", d.Endpoint), b)
	if err != nil {
		klog.Errorf("failed to create http request")
		return
	}
	request.URL.RawQuery = value.Encode()
	request.Header.Add("Content-Type", "application/json;charset=utf-8")
	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		klog.Errorf("failed to send msg to dingtalk. error: %s", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp != nil && resp.StatusCode != http.StatusOK {
		klog.Errorf("failed to send msg to dingtalk, because the response code is %d", resp.StatusCode)
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

func createMsgFromEvent(d *DingTalkSink, event *v1.Event) *DingTalkMsg {
	msg := &DingTalkMsg{}
	msg.MsgType = d.MsgType

	switch msg.MsgType {
	//https://open-doc.dingtalk.com/microapp/serverapi2/ye8tup#-6
	case MARKDOWN_MSG_TYPE:
		markdownCreator := NewMarkdownMsgBuilder(d.ClusterID, d.Region, event)
		markdownCreator.AddNodeName(event.Source.Host)
		markdownCreator.AddLabels(d.Labels)
		msg.Markdown = DingTalkMarkdown{
			//title 加不加其实没所谓,最终不会显示
			Title: fmt.Sprintf("Kubernetes(ID:%s) Event", d.ClusterID),
			Text:  markdownCreator.Build(),
		}
		break

	default:
		//默认按文本模式推送
		template := MSG_TEMPLATE
		if len(d.Labels) > 0 {
			for _, label := range d.Labels {
				template = fmt.Sprintf(LABEL_TEMPLATE, label) + template
			}
		}
		msg.Text = DingTalkText{
			Content: fmt.Sprintf(template, event.Type, event.InvolvedObject.Kind, event.Namespace, event.Name, event.Reason, util.GetLastEventTimestamp(event).Format(time.DateTime), event.Message),
		}
		break
	}

	return msg
}

func NewDingTalkSink(uri *url.URL) (*DingTalkSink, error) {
	d := &DingTalkSink{
		Level: WARNING,
	}
	if len(uri.Host) > 0 {
		d.Endpoint = uri.Host + uri.Path
	}
	opts := uri.Query()

	if len(opts["access_token"]) >= 1 {
		d.Token = opts["access_token"][0]
	} else {
		return nil, fmt.Errorf("you must provide dingtalk bot access_token")
	}

	if len(opts["level"]) >= 1 {
		d.Level = getLevel(opts["level"][0])
	}
	// get ding talk sign
	if len(opts["sign"]) >= 1 {
		d.Secret = opts["sign"][0]
	}
	//add extra labels
	if len(opts["label"]) >= 1 {
		d.Labels = opts["label"]
	}

	if msgType := opts["msg_type"]; len(msgType) >= 1 {
		d.MsgType = msgType[0]
	} else {
		//向下兼容,覆盖以前的版本,没有这个参数的情况
		d.MsgType = DEFAULT_MSG_TYPE
	}

	if clusterID := opts["cluster_id"]; len(clusterID) >= 1 {
		d.ClusterID = clusterID[0]
	}

	if region := opts["region"]; len(region) >= 1 {
		d.Region = region[0]
	}

	d.Namespaces = getValues(opts["namespaces"])
	// kinds:https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#lists-and-simple-kinds
	// such as node,pod,component and so on
	d.Kinds = getValues(opts["kinds"])

	return d, nil
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

func sign(t int64, secret string) string {
	strToHash := fmt.Sprintf("%d\n%s", t, secret)
	hmac256 := hmac.New(sha256.New, []byte(secret))
	hmac256.Write([]byte(strToHash))
	data := hmac256.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}
