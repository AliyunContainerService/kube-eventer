package eventbridge

import (
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/sinks/utils"
	"github.com/alibabacloud-go/eventbridge-sdk/eventbridge"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"math"
	"net/url"
	"time"
)

const (
	eventBridgeSinkName          = "EventBridgeSink"
	defaultBusName               = "default"
	eventBridgeEndpointSchema    = "%v.eventbridge.%v-vpc.aliyuncs.com"
	aliyunContainerServiceSource = "acs.cs"
	eventbridgeMaxBatchSize      = 16
)

type eventBridgeSink struct {
	client *eventbridge.Client
	akInfo *utils.AKInfo
}

func NewEventBridgeSink(uri *url.URL) (core.EventSink, error) {
	ebSink := &eventBridgeSink{}
	return ebSink, nil
}

func (ebSink *eventBridgeSink) Name() string {
	return eventBridgeSinkName
}

// Exports data to the external storage. The function should be synchronous/blocking and finish only
// after the given EventBatch was written. This will allow sink manager to push data only to these
// sinks that finished writing the previous data.
func (ebSink *eventBridgeSink) ExportEvents(batch *core.EventBatch) {
	if len(batch.Events) == 0 {
		return
	}

	batchSize := int(math.Ceil(float64(len(batch.Events)) / eventbridgeMaxBatchSize))
	for i := 0; i < batchSize; i++ {
		events := make([]*eventbridge.CloudEvent, eventbridgeMaxBatchSize)
		for j := i * batchSize; j < (i+1)*batchSize && j < len(batch.Events); j++ {
			cloudEvent, err := ebSink.toCloudEvent(batch.Events[j])
			if err != nil {
				klog.Errorf("failed to convert event %v to cloudevents, beacause of %v", batch.Events[j], err)
				continue
			}
			events = append(events, cloudEvent)
		}
		err := ebSink.putEvents(events)

		if err != nil {
			klog.Errorf("failed to put events to eventbridge, beacause of %v", err)
		}
	}
}

func (ebSink *eventBridgeSink) Stop() {
	//no background task, no need to implement
}

func (ebSink *eventBridgeSink) toCloudEvent(event *v1.Event) (*eventbridge.CloudEvent, error) {
	resourceName := event.Name
	kind := event.Kind
	namespace := event.Namespace
	subject := utils.CreateSelfLink(v1.ObjectReference{
		APIVersion: event.APIVersion,
		Kind:       kind,
		Name:       resourceName,
		Namespace:  namespace,
	})

	dataBytes, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	cloudEvent := new(eventbridge.CloudEvent).
		SetDatacontenttype("application/json").
		SetData(dataBytes).
		SetId(uuid.New().String()).
		SetSource(aliyunContainerServiceSource).
		SetTime(time.Now().String()).
		SetSubject(subject).
		SetType("type"). // TODO: Determine the appropriate event type
		SetExtensions(map[string]interface{}{
			"aliyuneventbusname": defaultBusName,
		})
	return cloudEvent, nil
}

func (ebSink *eventBridgeSink) putEvents(events []*eventbridge.CloudEvent) error {
	ebClient, err := ebSink.getClient()
	if err != nil {
		return err
	}
	_, err = ebClient.PutEvents(events)
	return err
}

func (ebSink *eventBridgeSink) getClient() (*eventbridge.Client, error) {
	if ebSink.client != nil && ebSink.isAkValid() {
		return ebSink.client, nil
	}
	return ebSink.newClient()
}

func (ebSink *eventBridgeSink) newClient() (*eventbridge.Client, error) {
	region, err := utils.ParseRegion()
	if err != nil {
		return nil, err
	}

	accountId, err := utils.ParseOwnerAccountId()
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(eventBridgeEndpointSchema, accountId, region)

	akInfo, err := utils.ParseAKInfo()
	if err != nil {
		return nil, err
	}

	config := &eventbridge.Config{}
	config.AccessKeyId = &akInfo.AccessKeyId
	config.AccessKeySecret = &akInfo.AccessKeySecret
	config.SecurityToken = &akInfo.SecurityToken
	config.Endpoint = &endpoint

	client, err := eventbridge.NewClient(config)
	if err != nil {
		return nil, err
	}

	ebSink.client = client
	ebSink.akInfo = akInfo

	return client, nil
}

func (ebSink *eventBridgeSink) isAkValid() bool {
	layout := "2006-01-02T15:04:05Z"
	t, err := time.Parse(layout, ebSink.akInfo.Expiration)
	if err != nil {
		klog.Errorf("failed to parse time layout, %v", err)
		return false
	}

	if t.Before(time.Now()) {
		klog.Error("invalid token which is expired")
		return false
	}

	t.Add(time.Minute * time.Duration(-10))
	if t.Before(time.Now()) {
		klog.Error("valid token which will be expired in ten minutes, should refresh it")
		return false
	}

	return true
}
