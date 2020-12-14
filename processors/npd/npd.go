package npd

import (
	"net/url"
	"regexp"

	"github.com/AliyunContainerService/kube-eventer/core"
	"k8s.io/apimachinery/pkg/types"
)

const podOOMKilling = "PodOOMKilling"

// pod was OOM killed. node:cn-hangzhou.172.16.xx.xx pod:xx-api-gateway-ack-deploy-xxxx-yyyy namespace:default uuid:123-ef51-46a5-b0a4-abc
var podOOMRegex = regexp.MustCompile(`node:(\S+)\s+pod:(\S+)\s+namespace:(\S+)\s+uuid:(\S+)`)

// EventProcessor is used to process node problem detector's events
type EventProcessor struct {
}

// Process convert PodOOMKilling event to pod event
func (ep *EventProcessor) Process(eb *core.EventBatch) *core.EventBatch {
	for _, event := range eb.Events {
		if event.Reason == podOOMKilling {
			event.InvolvedObject.Kind = "Pod"
			rst := podOOMRegex.FindStringSubmatch(event.Message)
			if len(rst) >= 5 {
				event.InvolvedObject.Name = rst[2]
				event.InvolvedObject.Namespace = rst[3]
				event.InvolvedObject.UID = types.UID(rst[4])

				event.ObjectMeta.Name = event.InvolvedObject.Name
				event.ObjectMeta.Namespace = event.InvolvedObject.Namespace
				event.ObjectMeta.UID = event.InvolvedObject.UID
			}
		}
	}
	return eb
}

// NewProcessor create a processor instance
func NewProcessor(url *url.URL) (core.EventProcessor, error) {
	return &EventProcessor{}, nil
}
