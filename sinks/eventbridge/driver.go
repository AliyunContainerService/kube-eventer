package eventbridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/AliyunContainerService/kube-eventer/sinks/utils"
	"github.com/AliyunContainerService/kube-eventer/util"
	"github.com/alibabacloud-go/eventbridge-sdk/eventbridge"
	ebUtil "github.com/alibabacloud-go/tea-utils/service"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	eventBridgeSinkName               = "EventBridgeSink"
	defaultBusName                    = "default"
	eventBridgeEndpointSchema         = "%v.eventbridge.%v.aliyuncs.com"
	eventBridgeInternalEndpointSchema = "%v.eventbridge.%v-vpc.aliyuncs.com"
	aliyunContainerServiceSource      = "acs.cs"
	eventbridgeMaxBatchSize           = 4
	eventTypeSchema                   = "cs:k8s:%vRelatedEvent"
	unknownEventType                  = "cs:k8s:UnknownTypeEvent"
)

type eventBridgeSink struct {
	client          *eventbridge.Client
	akInfo          *utils.AKInfo
	clusterId       string
	regionId        string
	accountId       string
	accessKeyId     string
	accessKeySecret string
	eventBusName    string
	internal        bool
}

type putEventsImpl func(events []*eventbridge.CloudEvent) error

func NewEventBridgeSink(uri *url.URL) (core.EventSink, error) {
	ebSink := &eventBridgeSink{}

	opts := uri.Query()

	if len(opts["clusterId"]) >= 1 {
		ebSink.clusterId = opts["clusterId"][0]
	} else {
		ebSink.clusterId = os.Getenv("ClusterId")
	}

	if len(ebSink.clusterId) == 0 {
		return nil, errors.New("please provide kubernetes cluster id via uri or env")
	}

	if len(opts["regionId"]) >= 1 {
		ebSink.regionId = opts["regionId"][0]
	} else {
		ebSink.regionId = os.Getenv("RegionId")
	}

	if len(opts["accessKeyId"]) >= 1 {
		ebSink.accessKeyId = opts["accessKeyId"][0]
	} else {
		ebSink.accessKeyId = os.Getenv("AccessKeyId")
	}

	if len(opts["accessKeySecret"]) >= 1 {
		ebSink.accessKeySecret = opts["accessKeySecret"][0]
	} else {
		ebSink.accessKeySecret = os.Getenv("AccessKeySecret")
	}

	if len(opts["eventBusName"]) >= 1 {
		ebSink.eventBusName = opts["eventBusName"][0]
	} else {
		ebSink.eventBusName = os.Getenv("EventBusName")
	}

	if len(ebSink.eventBusName) == 0 {
		ebSink.eventBusName = defaultBusName
	}

	accountId, err := utils.ParseOwnerAccountId()
	if err != nil {
		return nil, err
	}
	ebSink.accountId = accountId

	ebSink.internal = true
	if len(opts["internal"]) >= 1 {
		internal, err := strconv.ParseBool(opts["internal"][0])
		if err == nil {
			ebSink.internal = internal
		}
	}

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
	ebSink.exportEventsInBatch(batch, ebSink.putEvents)
}

func (ebSink *eventBridgeSink) Stop() {
	//no background task, no need to implement
}

func (ebSink *eventBridgeSink) toCloudEvent(event *v1.Event) (*eventbridge.CloudEvent, error) {
	subject := ebSink.createEventSubject(event.InvolvedObject)

	dataBytes, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	eType := unknownEventType
	if len(event.InvolvedObject.Kind) != 0 {
		eType = fmt.Sprintf(eventTypeSchema, event.InvolvedObject.Kind)
	}

	cloudEvent := new(eventbridge.CloudEvent).
		SetDatacontenttype("application/json").
		SetData(dataBytes).
		SetId(uuid.New().String()).
		SetSource(aliyunContainerServiceSource).
		SetTime(util.GetLastEventTimestamp(event).UTC().Format(time.RFC3339)).
		SetSubject(subject).
		SetType(eType).
		SetExtensions(map[string]interface{}{
			"aliyuneventbusname": ebSink.eventBusName,
		})
	return cloudEvent, nil
}

func (ebSink *eventBridgeSink) putEvents(events []*eventbridge.CloudEvent) error {
	ebClient, err := ebSink.getClient()
	if err != nil {
		return err
	}
	runtime := &ebUtil.RuntimeOptions{}
	runtime.SetAutoretry(true)
	_, err = ebClient.PutEventsWithOptions(events, runtime)
	return err
}

func (ebSink *eventBridgeSink) exportEventsInBatch(batch *core.EventBatch, putEvents putEventsImpl) {
	batchSize := int(math.Ceil(float64(len(batch.Events)) / eventbridgeMaxBatchSize))
	for i := 0; i < batchSize; i++ {
		events := make([]*eventbridge.CloudEvent, 0, eventbridgeMaxBatchSize)
		for j := i * eventbridgeMaxBatchSize; j < (i+1)*eventbridgeMaxBatchSize && j < len(batch.Events); j++ {
			cloudEvent, err := ebSink.toCloudEvent(batch.Events[j])
			if err != nil {
				klog.Errorf("failed to convert event %v to cloudevents, beacause of %v", batch.Events[j], err)
				continue
			}
			events = append(events, cloudEvent)
		}
		err := putEvents(events)

		if err != nil {
			klog.Errorf("failed to put events to eventbridge, beacause of %v", err)
		}
	}
}

func (ebSink *eventBridgeSink) getClient() (*eventbridge.Client, error) {
	if ebSink.client != nil && ebSink.isAkValid() {
		return ebSink.client, nil
	}
	return ebSink.newClient()
}

func (ebSink *eventBridgeSink) newClient() (*eventbridge.Client, error) {
	// region from env
	region, err := utils.GetRegionFromEnv()
	if err != nil {
		if ebSink.regionId != "" {
			// region from client
			region = ebSink.regionId
		} else {
			// region from meta data
			regionInMeta, err := utils.ParseRegionFromMeta()
			if err != nil {
				klog.Errorf("failed to get Region, because of %v", err)
				return nil, err
			}
			region = regionInMeta
		}
	}

	endpoint := ebSink.parseEventBridgeEndpoint(region)

	akInfo, err := utils.ParseAKInfoFromConfigPath()
	if err != nil {
		akInfo = &utils.AKInfo{}
		if ebSink.accessKeyId != "" && ebSink.accessKeySecret != "" {
			akInfo.AccessKeyId = ebSink.accessKeyId
			akInfo.AccessKeySecret = ebSink.accessKeySecret
		} else {
			akInfoInMeta, err := utils.ParseAKInfoFromMeta()
			if err != nil {
				klog.Errorf("failed to get RamRoleToken,because of %v", err)
				return nil, err
			}
			akInfo = akInfoInMeta
		}
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
	if len(ebSink.akInfo.Expiration) == 0 {
		return true
	}

	t, err := time.Parse(utils.StsTokenTimeLayout, ebSink.akInfo.Expiration)
	if err != nil {
		klog.Errorf("failed to parse time layout, %v", err)
		return false
	}
	nowT := time.Now()

	if t.Before(nowT) {
		klog.Error("invalid token which is expired")
		return false
	}

	t = t.Add(-time.Minute * time.Duration(10))
	if t.Before(nowT) {
		klog.Error("valid token which will be expired in ten minutes, should refresh it")
		return false
	}

	return true
}

// Creates a cloudevents subject of the form found in object metadata selfLinks
// like: acs:cs:${Region}:${Account}:${ClusterId}/${selfLink}
func (ebSink *eventBridgeSink) createEventSubject(o v1.ObjectReference) string {
	gvr, _ := meta.UnsafeGuessKindToResource(o.GroupVersionKind())
	versionNameHack := o.APIVersion

	// Core API types don't have a separate package name and only have a version string (e.g. /apis/v1/namespaces/default/pods/myPod)
	// To avoid weird looking strings like "v1/versionUnknown" we'll sniff for a "." in the version
	if strings.Contains(versionNameHack, ".") && !strings.Contains(versionNameHack, "/") {
		versionNameHack = versionNameHack + "/versionUnknown"
	}
	return fmt.Sprintf("acs:cs:%s:%s:%s/apis/%s/namespaces/%s/%s/%s", ebSink.regionId, ebSink.accountId,
		ebSink.clusterId, versionNameHack, o.Namespace, gvr.Resource, o.Name)
}

func (ebSink *eventBridgeSink) parseEventBridgeEndpoint(region string) string {
	if ebSink.internal {
		return fmt.Sprintf(eventBridgeInternalEndpointSchema, ebSink.accountId, region)
	}
	return fmt.Sprintf(eventBridgeEndpointSchema, ebSink.accountId, region)
}
