package eventbridge

import (
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/sinks/utils"
	"github.com/alibabacloud-go/eventbridge-sdk/eventbridge"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestNewEventBridgeSink(t *testing.T) {
	ebSink := createEventBridgeSink(t)
	assert.Equal(t, ebSink.accountId, "15210987")
	assert.Equal(t, ebSink.regionId, "cn-hangzhou")
	assert.Equal(t, ebSink.clusterId, "123")
}

func TestCreateEventSubject(t *testing.T) {
	ebSink := createEventBridgeSink(t)

	subject := ebSink.createEventSubject(v1.ObjectReference{
		APIVersion: "v120",
		Kind:       "pod",
		Name:       "my-pod",
		Namespace:  "my-namespace",
	})

	assert.Equal(t, "acs:cs:cn-hangzhou:15210987:123/apis/v120/namespaces/my-namespace/pods/my-pod", subject)
}

func TestToCloudEvent(t *testing.T) {
	ebSink := createEventBridgeSink(t)
	cloudEvent, err := ebSink.toCloudEvent(createTestEvent())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, *cloudEvent.Datacontenttype, "application/json")
	assert.Equal(t, cloudEvent.Extensions["aliyuneventbusname"], defaultBusName)
	assert.Equal(t, *cloudEvent.Source, aliyunContainerServiceSource)
	assert.Equal(t, *cloudEvent.Type, "cs:k8s:PodRelatedEvent")
	assert.Equal(t, *cloudEvent.Subject, "acs:cs:cn-hangzhou:15210987:123/apis/v1/namespaces/my-namespace/pods/my-pod")
}

func TestExportEventsInBatch(t *testing.T) {
	ebSink := createEventBridgeSink(t)
	batchEvents := &core.EventBatch{
		Timestamp: time.Now(),
	}
	var oneBatchEvents []*v1.Event
	for i := 0; i < eventbridgeMaxBatchSize; i++ {
		oneBatchEvents = append(oneBatchEvents, createTestEvent())
	}
	batchEvents.Events = oneBatchEvents

	ebSink.exportEventsInBatch(batchEvents, func(events []*eventbridge.CloudEvent) error {
		assert.Equal(t, len(events), eventbridgeMaxBatchSize)
		return nil
	})

	var twoBatchEvents []*v1.Event
	for i := 0; i < eventbridgeMaxBatchSize+2; i++ {
		twoBatchEvents = append(twoBatchEvents, createTestEvent())
	}
	batchEvents.Events = twoBatchEvents

	hitCnt := 0
	ebSink.exportEventsInBatch(batchEvents, func(events []*eventbridge.CloudEvent) error {
		if hitCnt == 0 {
			assert.Equal(t, len(events), eventbridgeMaxBatchSize)
			hitCnt++
		} else {
			assert.Equal(t, len(events), 2)
		}
		return nil
	})
}

func TestIsAkValid(t *testing.T) {
	ebSink := createEventBridgeSink(t)
	akInfo := utils.AKInfo{}
	ebSink.akInfo = &akInfo
	assert.Equal(t, ebSink.isAkValid(), true)

	expTime := time.Now()
	akInfo.Expiration = expTime.Add(time.Minute * time.Duration(15)).UTC().Format(utils.StsTokenTimeLayout)
	ebSink.akInfo = &akInfo
	assert.Equal(t, ebSink.isAkValid(), true)

	expTime = time.Now()
	akInfo.Expiration = expTime.Add(time.Minute * time.Duration(5)).UTC().Format(utils.StsTokenTimeLayout)
	ebSink.akInfo = &akInfo

	assert.Equal(t, ebSink.isAkValid(), false)

	expTime = time.Now()
	akInfo.Expiration = expTime.Add(time.Minute * time.Duration(-5)).UTC().Format(utils.StsTokenTimeLayout)
	ebSink.akInfo = &akInfo

	assert.Equal(t, ebSink.isAkValid(), false)
}

func createEventBridgeSink(t *testing.T) *eventBridgeSink {
	os.Setenv("RegionId", "cn-hangzhou")
	os.Setenv("OwnerAccountId", "15210987")

	uri := url.URL{
		RawQuery: "clusterId=123",
	}
	sink, err := NewEventBridgeSink(&uri)

	if err != nil {
		t.Fatal(err)
	}

	ebSink := sink.(*eventBridgeSink)
	return ebSink
}

func createTestEvent() *v1.Event {
	now := time.Now()
	event := &v1.Event{
		Reason:         "A warning event occurs",
		Message:        "some node emits the event without empty message",
		Count:          251,
		LastTimestamp:  metav1.NewTime(now),
		FirstTimestamp: metav1.NewTime(now),
		Type:           "Warning",
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-event",
			Namespace: "my-namespace",
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Event",
		},
		InvolvedObject: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       "my-pod",
			Namespace:  "my-namespace",
		},
	}
	return event
}
