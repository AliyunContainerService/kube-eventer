// Copyright 2018 Google Inc. All Rights Reserved.
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

package cms

import (
	"encoding/json"
	"errors"
	"github.com/AliyunContainerService/kube-eventer/core"
	. "github.com/AliyunContainerService/kube-eventer/sinks/cms/metrichub/v20211001"
	"github.com/AliyunContainerService/kube-eventer/sinks/utils"
	k8s "k8s.io/api/core/v1"
	"k8s.io/klog"
	"math"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultBufferSize = 512
)

// SysEventRing 环形队列，循环写入不阻塞，读阻塞。堆积时最老的数据被替换
type SysEventRing struct {
	cond     *sync.Cond
	popCount int // 累计发送的数量(含discard的数量)
	count    int
	buf      []*SystemEvent // 最大缓存4K，再多就丢掉旧数据
	close    bool
}

func (p *SysEventRing) initialize(capacity int) {
	p.cond = sync.NewCond(new(sync.Mutex))
	p.buf = make([]*SystemEvent, capacity)
	p.count = 0
	p.popCount = 0
	p.close = false
}

func (p *SysEventRing) phyIndex(n int) int {
	return n % len(p.buf)
}

func (p *SysEventRing) push(events []*SystemEvent) (r int) {
	tryPush := func(events []*SystemEvent) (r []*SystemEvent, ok bool) {
		p.cond.L.Lock()
		defer p.cond.L.Unlock()

		if ok = !p.close; ok {
			empty := p.popCount == p.count // 队列是否为空

			begin := p.phyIndex(p.count)
			count := copy(p.buf[begin:], events)
			p.count += count
			r = events[count:]

			if empty {
				p.cond.Signal()
			}
		}
		return
	}

	for ok := true; ok && len(events) > 0; r++ {
		events, ok = tryPush(events)
	}
	return
}

func (p *SysEventRing) front(maxCount int) (r []*SystemEvent, discard int, ok bool) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	for p.popCount == p.count {
		if p.close {
			return
		} else {
			p.cond.Wait() // 列表为空，则等待
		}
	}
	ok = true

	begin := p.popCount
	if p.count > p.popCount+len(p.buf) {
		begin = p.count - len(p.buf)
	}
	discard = begin - p.popCount
	p.popCount = begin

	count := p.count - begin
	if count > maxCount {
		count = maxCount
	}

	for count > 0 {
		offset := p.phyIndex(begin)
		end := offset + count
		if end > len(p.buf) {
			end = len(p.buf)
		}
		r = append(r, p.buf[offset:end]...)
		count -= end - offset
		begin = end
	}

	return
}

func (p *SysEventRing) pop(count int) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.popCount += count

	// 防溢出
	if p.count >= math.MaxUint16 {
		diff := p.popCount - p.phyIndex(p.popCount)
		p.popCount -= diff
		p.count -= diff
	}
}

func (p *SysEventRing) Stop() {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	p.close = true
	p.cond.Broadcast()
}

type IClient interface {
	PutSystemEvent(events []*SystemEvent) (response PutSystemEventResponse, err error)
}

// CmsSink aliyun cloud monitor
type tagCmsSink struct {
	// 环式缓存
	SysEventRing

	regionId string
	level    string
	client   IClient
}

func (*tagCmsSink) Name() string {
	return "CmsSink"
}

func (d *tagCmsSink) ExportEvents(batch *core.EventBatch) {
	if batch != nil && len(batch.Events) > 0 {
		events := make([]*SystemEvent, 0, len(batch.Events))
		for _, coreEvent := range batch.Events {
			if event := d.ConvertToSysEvent(coreEvent); event != nil {
				events = append(events, event)
			}
		}
		d.push(events)
	}
}

func (d *tagCmsSink) loopConsume(chanClose chan struct{}, batchCount int) {
	defer close(chanClose)

	totalDiscard := 0
	for {
		if events, discard, ok := d.front(batchCount); !ok {
			break
		} else {
			totalDiscard += discard
			if totalDiscard > 0 {
				var content map[string]interface{}
				if err := json.Unmarshal([]byte(events[0].Content), &content); err == nil {
					content["discard"] = totalDiscard
					if jsonBytes, err := json.Marshal(content); err == nil {
						events[0].Content = string(jsonBytes)
					}
				}
			}
			if _, err := d.client.PutSystemEvent(events); err == nil {
				totalDiscard = 0
				d.pop(len(events))
			} // else // 此次作废，下次重发
		}
	}
}

func getEventTime(event *k8s.Event, now func() time.Time) string {
	var eventTime time.Time

	switch {
	case !event.LastTimestamp.IsZero():
		eventTime = event.LastTimestamp.Time
	case !event.EventTime.IsZero():
		eventTime = event.EventTime.Time
	default:
		eventTime = now()
	}

	return eventTime.Format(EventTimeLayout)
}

/*
{
    "message": "MountVolume.SetUp failed for volume \"eventer-token\" : secret \"addon.log.token\" not found",
    "reportingInstance": "",
    "count": 82,
    "source": {
        "host": "cn-qingdao.172.28.103.239",
        "component": "kubelet"
    },
    "reason": "FailedMount",
    "type": "Warning",  // 只有Normal和Warning
    "reportingComponent": "",
    "lastTimestamp": "2022-04-14T10:00:48Z",
    "firstTimestamp": "2022-04-14T07:30:10Z",
    "involvedObject": {
        "apiVersion": "v1",
        "uid": "8988b623-00cc-4f2b-be36-3286240ab95b",
        "resourceVersion": "3405",
        "name": "ack-node-problem-detector-eventer-598b6bf66b-m2n7p",
        "kind": "Pod",
        "namespace": "kube-system"
    },
    "metadata": {
        "uid": "ea76e1cd-9849-4587-9226-968b729512d6",
        "resourceVersion": "48452",
        "name": "ack-node-problem-detector-eventer-598b6bf66b-m2n7p.16e5b2c809e990cf",
        "managedFields": [
            {
                "apiVersion": "v1",
                "operation": "Update",
                "time": "2022-04-14T07:30:10Z",
                "manager": "kubelet",
                "fieldsType": "FieldsV1",
                "fieldsV1": {
                    "f:source": {
                        "f:host": {},
                        "f:component": {}
                    },
                    "f:firstTimestamp": {},
                    "f:count": {},
                    "f:involvedObject": {
                        "f:kind": {},
                        "f:uid": {},
                        "f:name": {},
                        "f:apiVersion": {},
                        "f:namespace": {},
                        "f:resourceVersion": {}
                    },
                    "f:type": {},
                    "f:reason": {},
                    "f:message": {},
                    "f:lastTimestamp": {}
                }
            }
        ],
        "creationTimestamp": "2022-04-14T07:30:10Z",
        "namespace": "kube-system"
    }
}
*/

func (d *tagCmsSink) ConvertToSysEvent(event *k8s.Event) (r *SystemEvent) {
	r = &SystemEvent{
		Product:   Product,
		EventType: event.Reason,
		Name:      event.GetName(),
		EventTime: getEventTime(event, time.Now),
		GroupId:   "0", // 跟昱杰沟通，此处先填0，以后如果需要groupId，再升级插件。
		Resource:  string(event.InvolvedObject.UID),
		// ResourceId: "acs:" + Product + ":" + d.regionId + "::uuid/" + string(event.InvolvedObject.UID),
		Level:  d.level,
		Status: event.Type,
		// UserId:     "",
		// Tags:       "",
		RegionId: d.regionId,
		Time:     time.Now().Format(EventTimeLayout),
	}

	if jsonBytes, err := json.Marshal(event); err == nil {
		r.Content = string(jsonBytes)
	}

	return
}

type Config struct {
	endPoint        string
	accessKeyId     string
	accessKeySecret string
	regionId        string
	level           string
}

// ParseConfig create config from uri
func ParseConfig(uri *url.URL) *Config {
	c := &Config{}

	opts := uri.Query()
	if uri.Host != "" && !strings.EqualFold(uri.Host, "unknown") {
		c.endPoint = uri.Scheme + "://" + uri.Host
	}

	doGet := func(optKey, envKey string, def string) string {
		if len(opts[optKey]) >= 1 {
			return opts[optKey][0]
		}
		if envKey != "" {
			return os.Getenv(envKey)
		}
		return def
	}

	c.regionId = doGet("regionId", "RegionId", "")
	c.accessKeyId = doGet("accessKeyId", "AccessKeyId", "")
	c.accessKeySecret = doGet("accessKeySecret", "AccessKeySecret", "")
	c.level = doGet("level", "", "INFO")

	if c.endPoint == "" {
		endPoint := GetEndPoint(c.regionId)
		c.endPoint = endPoint.EndPoints[0]
		if c.regionId == "" {
			c.regionId = endPoint.RegionId
		}
	}
	if c.regionId == "" {
		c.regionId = DefaultRegionId
	}

	return c
}

func GetRegion(c *Config, fnParseRegionFromMeta func() (string, error)) (regionId string, err error) {
	if c != nil && c.regionId != "" {
		// region from client
		regionId = c.regionId
	} else if regionId, err = fnParseRegionFromMeta(); err != nil {
		// region from meta data
		klog.Errorf("failed to get Region,because of %v", err)
	}
	return
}

func GetAKInfo(c *Config) (*utils.AKInfo, error) {
	// 1. first get ak/sk from env
	if c != nil && c.accessKeyId != "" && c.accessKeySecret != "" {
		return &utils.AKInfo{
			AccessKeyId:     c.accessKeyId,
			AccessKeySecret: c.accessKeySecret,
		}, nil
	}

	// 2. aliyun akInfo fetch
	if akInfo := fnGetAkInfo(utils.CMSConfigPath); akInfo != nil {
		return akInfo, nil
	}

	klog.Errorf("get sls akInfo error.")
	return nil, errors.New("get cms akInfo error")
}

var (
	fnGetAkInfo = utils.GetAkInfo
)

func newClient(c *Config) (client *Client, err error) {
	// // get region from env
	// region, err := utils.GetRegionFromEnv()
	// if err != nil {
	// 	if c.regionId != "" {
	// 		// region from client
	// 		region = c.regionId
	// 	} else {
	// 		// region from meta data
	// 		region, err = utils.ParseRegionFromMeta()
	// 		if err != nil {
	// 			klog.Errorf("failed to get Region,because of %v", err)
	// 			return
	// 		}
	// 	}
	// 	err = nil
	// }

	var akInfo *utils.AKInfo
	if akInfo, err = GetAKInfo(c); err == nil {
		client = CreateMetricHubClient(c.endPoint, akInfo.AccessKeyId, akInfo.AccessKeySecret, akInfo.SecurityToken)
	}
	return
}

type tagCmsSinkOpt struct {
	bufferSize int
	startLoop  bool
}

// NewCmsSink Usage:
// --sink=cms:http://metrichub-[your_region_id].aliyun-inc.com?regionId=[your_region_id]&accessKeyId=[your_access_key]&accessKeySecret=[you_access_secret]&level=[alert_level]
func NewCmsSink(uri *url.URL, opt *tagCmsSinkOpt) (r core.EventSink, err error) {
	c := ParseConfig(uri)

	sink := &tagCmsSink{level: c.level}
	sink.regionId, err = GetRegion(c, utils.ParseRegionFromMeta)
	if err == nil {
		var client *Client
		if client, err = newClient(c); err == nil {
			sink.client = client
		}
	}
	if err == nil {
		if opt == nil {
			opt = &tagCmsSinkOpt{bufferSize: defaultBufferSize, startLoop: true}
		}
		sink.SysEventRing.initialize(opt.bufferSize)
		if opt.startLoop {
			const batchCount = 10 // 一次最多发送10个事件，避免过多时超限
			go sink.loopConsume(make(chan struct{}), batchCount)
		}
		r = sink
	}

	return
}
