package cms

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/core"
	. "github.com/AliyunContainerService/kube-eventer/sinks/cms/metrichub/v20211001"
	"github.com/AliyunContainerService/kube-eventer/sinks/utils"
	"github.com/stretchr/testify/assert"
	k8s "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"
)

const bufferCapacity = defaultBufferSize

func TestSysEventRing(t *testing.T) {
	ring := SysEventRing{}
	ring.initialize(bufferCapacity)
	assert.Equal(t, bufferCapacity, len(ring.buf))
	assert.Zero(t, ring.popCount)
	assert.Zero(t, ring.count)
	assert.NotNil(t, ring.count)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ring.push([]*SystemEvent{{Product: "acs_container"}})
	}()

	events, discard, ok := ring.front(BatchCount)
	assert.True(t, ok)
	assert.Zero(t, discard)
	assert.Equal(t, 1, len(events))
	assert.Equal(t, "acs_container", events[0].Product)

	assert.Zero(t, ring.popCount)
	assert.Equal(t, 1, ring.count)

	ring.pop(len(events))
	assert.Equal(t, ring.popCount, ring.count)

	wg.Wait()
}

func TestSysEventRing_Overflow(t *testing.T) {
	ring := SysEventRing{}
	ring.initialize(bufferCapacity)
	assert.Equal(t, bufferCapacity, len(ring.buf))
	assert.Zero(t, ring.popCount)
	assert.Zero(t, ring.count)
	assert.False(t, ring.close)

	const overflow = 32
	expectCount := math.MaxUint16 + overflow
	for i := 0; i < expectCount; i++ {
		count := ring.push([]*SystemEvent{
			{
				Product: "acs_container",
				Name:    fmt.Sprintf("test.%v", i),
			},
		})
		assert.Equal(t, 1, count)
	}
	assert.Zero(t, ring.popCount)
	assert.Equal(t, expectCount, ring.count)

	for offset := 0; offset < bufferCapacity; offset += BatchCount {
		events, discard, ok := ring.front(BatchCount)
		assert.True(t, ok)
		if offset == 0 {
			assert.Equal(t, expectCount-bufferCapacity, discard)
		} else {
			assert.Zero(t, discard)
		}
		expectEventCount := BatchCount
		if offset+BatchCount > bufferCapacity {
			expectEventCount = bufferCapacity - offset
		}
		assert.Equal(t, expectEventCount, len(events))
		for i, event := range events {
			assert.Equal(t, "acs_container", event.Product)
			expectName := fmt.Sprintf("test.%v", expectCount-bufferCapacity+offset+i)
			if expectName != event.Name {
				fmt.Printf("event[%v].Name ==> %v, expect: %v\n", i+offset, event.Name, expectName)
			}
			assert.Equal(t, expectName, event.Name)
		}

		if offset == 0 {
			assert.Equal(t, expectCount-bufferCapacity, ring.popCount)
			assert.Equal(t, expectCount, ring.count)
		}

		ring.pop(len(events))
		assert.Equal(t, ring.popCount, ring.count-bufferCapacity+offset+len(events))
		assert.Less(t, ring.count, 2*bufferCapacity)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		ring.Stop()
	}()
	events, discard, ok := ring.front(BatchCount) // ring关闭，返回无效数据
	assert.Nil(t, events)
	assert.Zero(t, discard)
	assert.False(t, ok)

	wg.Wait()
}

func TestNewCmsSing(t *testing.T) {
	defer func(val string) { _ = os.Setenv("RegionId", val) }(os.Getenv("RegionId"))
	_ = os.Setenv("RegionId", "cn-beijing")
	params, err := url.Parse(GetEndPoint("").EndPoints[0] + `?accessKeyId=1&accessKeySecret=2`)
	assert.NoError(t, err)
	sink, err := NewCmsSink(params, &tagCmsSinkOpt{bufferSize: bufferCapacity, startLoop: false})
	assert.NoError(t, err)
	defer sink.Stop()
	assert.Equal(t, "CmsSink", sink.Name())
	cmsSink := sink.(*tagCmsSink)
	assert.NotNil(t, cmsSink.buf)
	assert.Equal(t, "cn-beijing", cmsSink.config.regionId)
}

func TestGetEnvTime(t *testing.T) {
	now := time.Now()
	expectTime := now.Format(EventTimeLayout)

	assert.Equal(t, expectTime, getEventTime(&k8s.Event{LastTimestamp: metav1.NewTime(now)}, nil))
	assert.Equal(t, expectTime, getEventTime(&k8s.Event{EventTime: metav1.NewMicroTime(now)}, nil))
	assert.Equal(t, expectTime, getEventTime(&k8s.Event{}, func() time.Time { return now }))
}

type tagFakeClient struct {
	ch chan []*SystemEvent
}

func (d *tagFakeClient) PutSystemEvent(events []*SystemEvent) (response PutSystemEventResponse, err error) {
	select {
	case d.ch <- events:
	default:
	}
	return
}

func TestTagCmsSink_ExportEvents(t *testing.T) {
	params, err := url.Parse(GetEndPoint("").EndPoints[0] + `?regionId=cn-zhangjiakou&accessKeyId=1&accessKeySecret=2`)
	assert.NoError(t, err)
	sink, err := NewCmsSink(params, nil)
	assert.NoError(t, err)
	defer sink.Stop()
	assert.Equal(t, "CmsSink", sink.Name())
	cmsSink := sink.(*tagCmsSink)
	fakeClient := tagFakeClient{ch: make(chan []*SystemEvent, 1)}
	cmsSink.client = &fakeClient

	now := time.Now()
	event := k8s.Event{
		LastTimestamp: metav1.NewTime(now),
	}
	cmsSink.ExportEvents(&core.EventBatch{Events: []*k8s.Event{&event}})
	events := <-fakeClient.ch
	assert.Equal(t, 1, len(events))
	assert.Equal(t, Product, events[0].Product)
	assert.NotEmpty(t, events[0].Content)
}

func TestTagCmsSink_ExportEvents_DiscardNotZero(t *testing.T) {
	params, err := url.Parse(GetEndPoint("").EndPoints[0] + `?regionId=cn-zhangjiakou&accessKeyId=1&accessKeySecret=2`)
	assert.NoError(t, err)
	sink, err := NewCmsSink(params, &tagCmsSinkOpt{bufferSize: bufferCapacity, startLoop: false})
	assert.NoError(t, err)
	assert.Equal(t, "CmsSink", sink.Name())
	cmsSink := sink.(*tagCmsSink)
	fakeClient := tagFakeClient{ch: make(chan []*SystemEvent, 2)}
	cmsSink.client = &fakeClient

	now := time.Now()
	event := k8s.Event{
		LastTimestamp: metav1.NewTime(now),
	}
	const expectDiscard = 32
	for i := 0; i < bufferCapacity+expectDiscard; i++ {
		cmsSink.ExportEvents(&core.EventBatch{Events: []*k8s.Event{&event}})
	}
	chanClose := make(chan struct{})
	go cmsSink.loopConsume(chanClose, BatchCount)

	events := <-fakeClient.ch
	assert.Equal(t, BatchCount, len(events))
	assert.Equal(t, Product, events[0].Product)
	fmt.Println("Content:", events[0].Content)
	assert.NotEmpty(t, events[0].Content)
	assert.NotEmpty(t, events[1].Content)
	content := struct {
		Discard int `json:"discard"`
	}{}
	assert.NoError(t, json.Unmarshal([]byte(events[0].Content), &content))
	assert.Equal(t, expectDiscard, content.Discard)

	cmsSink.Stop()
	<-chanClose // 确保loopConsumer已关闭
}

func TestGetRegion(t *testing.T) {
	err := errors.New("ByDesign")
	regionId, actualErr := GetRegion(nil, func() (string, error) {
		return "", err
	})
	assert.Empty(t, regionId)
	assert.Same(t, err, actualErr)
}

func TestGetAkInfo(t *testing.T) {
	defer func(old func(configPath string) *utils.AKInfo) { fnGetAkInfo = old }(fnGetAkInfo)
	fakeAkInfo := &utils.AKInfo{}
	fnGetAkInfo = func(string) *utils.AKInfo {
		return fakeAkInfo
	}
	akInfo, err := GetAKInfo(nil)
	assert.NoError(t, err)
	assert.Same(t, fakeAkInfo, akInfo)
}

// 这个会阻塞
func TestNewClient_Error(t *testing.T) {
	defer func(old func(configPath string) *utils.AKInfo) { fnGetAkInfo = old }(fnGetAkInfo)
	fnGetAkInfo = func(string) *utils.AKInfo { return nil }
	client, err := newClient(nil)
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestParseConfig(t *testing.T) {
	params, err := url.Parse(`https://unknown?regionId=cn-hangzhou&userId=123&accessKeyId=1&accessKeySecret=2`)
	assert.NoError(t, err)
	c, _ := ParseConfig(params)
	assert.Contains(t, c.endPoint, "://")
	assert.Contains(t, c.endPoint, "cn-hangzhou")
	assert.Equal(t, "cn-hangzhou", c.regionId)
}

func TestParseConfig2(t *testing.T) {
	params, err := url.Parse(`https://host.com/?regionId=cn-hangzhou&userId=123accessKeyId=1&accessKeySecret=2`)
	assert.NoError(t, err)
	c, _ := ParseConfig(params)
	assert.Equal(t, c.endPoint, `https://host.com`)
	assert.Equal(t, "cn-hangzhou", c.regionId)
}

func TestParseConfig3(t *testing.T) {
	params, err := url.Parse(`?regionId=cn-hangzhou&userId=123accessKeyId=1&accessKeySecret=2`)
	assert.NoError(t, err)
	c, _ := ParseConfig(params)
	assert.Contains(t, c.endPoint, "cn-hangzhou")
	assert.Equal(t, DefaultRegionId, c.regionId)
}
