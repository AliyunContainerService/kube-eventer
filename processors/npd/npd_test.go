package npd

import (
	"fmt"
	"testing"
	"time"

	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/stretchr/testify/assert"
	kube_api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFlow(t *testing.T) {

	processor, _ := NewProcessor(nil)

	batch := &core.EventBatch{
		Timestamp: time.Now(),
		Events: []*kube_api.Event{
			&kube_api.Event{
				InvolvedObject: kube_api.ObjectReference{
					Kind: "Node",
					Name: "cn-beijing.10.1.1.1",
					UID:  "cn-beijing.10.1.1.1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cn-beijing.10.1.1.1",
					Namespace: "d",
				},
				Reason:  podOOMKilling,
				Message: "pod was OOM killed. node:cn-hangzhou.172.16.xx.xx pod:xx-api-gateway-ack-deploy-xxxx-yyyy namespace:ack uuid:123-ef51-46a5-b0a4-abc",
			},
		},
	}

	batch = processor.Process(batch)

	assert.Equal(t, len(batch.Events), 1)
	event := batch.Events[0]
	assert.Equal(t, event.InvolvedObject.Kind, "Pod")
	assert.Equal(t, event.InvolvedObject.Name, "xx-api-gateway-ack-deploy-xxxx-yyyy")
	assert.Equal(t, event.InvolvedObject.Namespace, "ack")
	assert.Equal(t, fmt.Sprintln(event.InvolvedObject.UID), fmt.Sprintln("123-ef51-46a5-b0a4-abc"))
	assert.Equal(t, event.InvolvedObject.UID, event.ObjectMeta.UID)
	assert.Equal(t, event.InvolvedObject.Name, event.ObjectMeta.Name)
	assert.Equal(t, event.InvolvedObject.Namespace, event.ObjectMeta.Namespace)
}
