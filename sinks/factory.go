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

package sinks

import (
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/common/flags"
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/sinks/dingtalk"
	"github.com/AliyunContainerService/kube-eventer/sinks/elasticsearch"
	"github.com/AliyunContainerService/kube-eventer/sinks/eventbridge"
	"github.com/AliyunContainerService/kube-eventer/sinks/honeycomb"
	"github.com/AliyunContainerService/kube-eventer/sinks/influxdb"
	"github.com/AliyunContainerService/kube-eventer/sinks/kafka"
	"github.com/AliyunContainerService/kube-eventer/sinks/log"
	"github.com/AliyunContainerService/kube-eventer/sinks/mongo"
	"github.com/AliyunContainerService/kube-eventer/sinks/mysql"
	"github.com/AliyunContainerService/kube-eventer/sinks/pulsar"
	"github.com/AliyunContainerService/kube-eventer/sinks/riemann"
	"github.com/AliyunContainerService/kube-eventer/sinks/sls"
	"github.com/AliyunContainerService/kube-eventer/sinks/webhook"
	"github.com/AliyunContainerService/kube-eventer/sinks/wechat"
	"k8s.io/klog"
)

type SinkFactory struct {
}

func (this *SinkFactory) Build(uri flags.Uri) (core.EventSink, error) {
	switch uri.Key {
	case "log":
		return logsink.CreateLogSink()
	case "influxdb":
		return influxdb.CreateInfluxdbSink(&uri.Val)
	case "mysql":
		return mysql.CreateMysqlSink(&uri.Val)
	case "elasticsearch":
		return elasticsearch.NewElasticSearchSink(&uri.Val)
	case "kafka":
		return kafka.NewKafkaSink(&uri.Val)
	case "riemann":
		return riemann.CreateRiemannSink(&uri.Val)
	case "honeycomb":
		return honeycomb.NewHoneycombSink(&uri.Val)
	case "dingtalk":
		return dingtalk.NewDingTalkSink(&uri.Val)
	case "sls":
		return sls.NewSLSSink(&uri.Val)
	case "wechat":
		return wechat.NewWechatSink(&uri.Val)
	case "webhook":
		return webhook.NewWebHookSink(&uri.Val)
	case "eventbridge":
		return eventbridge.NewEventBridgeSink(&uri.Val)
	case "mongo":
		return mongo.CreateMongoSink(&uri.Val)
	case "pulsar":
		return pulsar.NewPulsarSink(&uri.Val)
	default:
		return nil, fmt.Errorf("Sink not recognized: %s", uri.Key)
	}
}

func (this *SinkFactory) BuildAll(uris flags.Uris) []core.EventSink {
	result := make([]core.EventSink, 0, len(uris))
	for _, uri := range uris {
		sink, err := this.Build(uri)
		if err != nil {
			klog.Errorf("Failed to create %v sink: %v", uri, err)
			continue
		}
		result = append(result, sink)
	}
	return result
}

func NewSinkFactory() *SinkFactory {
	return &SinkFactory{}
}
