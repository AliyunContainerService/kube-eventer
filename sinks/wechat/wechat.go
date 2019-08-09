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

package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AliyunContainerService/kube-eventer/core"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	WECHAT_SINK         = "WechatSink"
	WARNING           int = 2
	NORMAL            int = 1
	DEFAULT_MSG_TYPE      = "text"
	CONTENT_TYPE_JSON     = "application/json"
	LABE_TEMPLATE         = "%s\n"
	//发送消息使用导的url
	sendurl = `https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=`
	//获取token使用导的url
	get_token = `https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=`
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

/**
wechat msg struct
*/
type WechatMsg struct {
	ToUser   string      `json:"touser"`
	ToParty  string      `json:"toparty"`
	ToTag    string      `json:"totag"`
	MsgType  string      `json:"msgtype"`
	AgentID  int         `json:"agentid"`
	Text     WechatText  `json:"text"`
	Safe     int         `json:"safe"`
}

type WechatText struct {
	Content string `json:"content"`
}

/**
dingtalk sink usage
--sink:wechat:https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=[access_token]&level=Warning&label=[label]

level: Normal or Warning. The event level greater than global level will emit.
label: some thing unique when you want to distinguish different k8s clusters.
*/
type WechatSink struct {
	Endpoint   string
	Namespaces []string
	Kinds      []string
	CorpID     string
	CorpSecret string
	AgentID    int
	ToUser     []string
	Level      int
	Labels     []string
	MsgType    string
	ClusterID  string
	Region     string
}

func (d *WechatSink) Name() string {
	return WECHAT_SINK
}

func (d *WechatSink) Stop() {
	//do nothing
}

func (d *WechatSink) ExportEvents(batch *core.EventBatch) {
	for _, event := range batch.Events {
		if d.isEventLevelDangerous(event.Type) {
			d.Send(event)
			// add threshold
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (d *WechatSink) isEventLevelDangerous(level string) bool {
	score := getLevel(level)
	if score >= d.Level {
		return true
	}
	return false
}

func (d *WechatSink) Send(event *v1.Event) {
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

	token, err := getToken(d.CorpID, d.CorpSecret)
	klog.Error(token)

	if err != nil {
		klog.Warningf("failed to get token,because of %v", err)
		return
	}

	for _, user := range d.ToUser {
		msg.ToUser = user
		msg_bytes, err := json.Marshal(msg)
		if err != nil {
			klog.Warningf("failed to marshal msg %v", msg)
			return
		}

		b := bytes.NewBuffer(msg_bytes)
		resp, err := http.Post(sendurl+token, CONTENT_TYPE_JSON, b)
		if err != nil {
			klog.Errorf("failed to send msg to dingtalk. error: %s", err.Error())
			return
		}
		if resp != nil && resp.StatusCode != http.StatusOK {
			klog.Errorf("failed to send msg to dingtalk, because the response code is %d", resp.StatusCode)
			return
		}
	}

}

func getToken(corp_id, corp_secret string) (string, error)  {
	resp, err := http.Get(get_token + corp_id + "&corpsecret=" + corp_secret)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("get wechat token request error")
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	var token string
	err = json.Unmarshal(buf, &token)
	return token, nil
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

func createMsgFromEvent(d *WechatSink, event *v1.Event) *WechatMsg {
	msg := &WechatMsg{}
	msg.MsgType = d.MsgType
	msg.AgentID = d.AgentID

	//默认按文本模式推送
	template := MSG_TEMPLATE
	if len(d.Labels) > 0 {
		for _, label := range d.Labels {
			template = fmt.Sprintf(LABE_TEMPLATE, label) + template
		}
	}

	msg.Text = WechatText{
		Content: fmt.Sprintf(template, event.Type, event.InvolvedObject.Kind, event.Namespace, event.Name, event.Reason, event.LastTimestamp.String(), event.Message),
	}

	return msg
}


func NewWechatSink(uri *url.URL) (*WechatSink, error) {
	d := &WechatSink{
		Level: WARNING,
	}
	opts := uri.Query()

	if len(opts["corp_id"]) >= 1 {
		d.CorpID = opts["corp_id"][0]
	} else {
		return nil, fmt.Errorf("you must provide wechat corpid")
	}

	if len(opts["corp_secret"]) >= 1 {
		d.CorpSecret = opts["corp_secret"][0]
	} else {
		return nil, fmt.Errorf("you must provide wechat corpsecret")
	}

	if len(opts["agent_id"]) >= 1 {
		if AgentID, err:= strconv.Atoi(opts["agent_id"][0]); err == nil {
			d.AgentID = AgentID
		} else {
			return nil, fmt.Errorf("you must provide wechat agentid is number")
		}
	} else {
		return nil, fmt.Errorf("you must provide wechat agentid")
	}

	if len(opts["to_user"]) >= 1 && opts["to_user"][0] != "" {
		for _, user := range strings.Split(opts["to_user"][0],",") {
			d.ToUser = append(d.ToUser, user)
		}
	} else {
		d.ToUser = append(d.ToUser, "@all")
	}

	if len(opts["level"]) >= 1 {
		d.Level = getLevel(opts["level"][0])
	}

	//add extra labels
	if len(opts["label"]) >= 1 {
		d.Labels = opts["label"]
	}

	if msgType := opts["msg_type"]; len(msgType) >= 1 {
		d.MsgType = msgType[0]
	} else {
		d.MsgType = DEFAULT_MSG_TYPE
	}

	if clusterID := opts["cluster_id"]; len(clusterID) >= 1 {
		d.ClusterID = clusterID[0]
	}

	if region := opts["region"]; len(region) >= 1 {
		d.Region = region[0]
	}

	d.Namespaces = getValues(opts["namespaces"])
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
