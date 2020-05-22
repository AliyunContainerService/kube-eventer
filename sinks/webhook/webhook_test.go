package webhook

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	webhookSinkValid  = "https://oapi.dingtalk.com/robot/send?access_token=token&level=Normal&namespaces=a,b&kinds=c,d&header=contentType=demo&header=content2=3"
	webhookSinkFilter = "https://oapi.example.com/robot/send?access_token=token&level=Normal&namespaces=a,b&kinds=c,d&reason=Failed&header=contentType=demo&header=content2=3"
)

var (
	kubeSystemNamespaceEventPod = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Event2",
			Namespace: "kube-system",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Reason:  "FailedStartUp",
		Type:    Normal,
		Message: "DEMO",
	}

	newWebhookFilterEventPod = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Event1",
			Namespace: "a",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "c",
		},
		Reason:  "FailedStartUp",
		Type:    Normal,
		Message: "DEMO",
	}
)

func TestNewWebhookSink(t *testing.T) {
	uri, err := url.Parse(webhookSinkValid)
	if err != nil {
		t.Fatalf("Failed to prase webhookSinkValid,err: %v", err)
	}
	w, err := NewWebHookSink(uri)
	if err != nil {
		t.Fatalf("Failed to create NewWebhookSink,err: %v", err)
	}

	assert.True(t, webhookSinkValid == w.endpoint, "endpoint should be the same")
}

func TestNewWebhookFilter(t *testing.T) {
	uri, err := url.Parse(webhookSinkFilter)
	if err != nil {
		t.Fatalf("Failed to prase webhookSinkFilter,err: %v", err)
	}
	w, err := NewWebHookSink(uri)
	if err != nil {
		t.Fatalf("Failed to create NewWebhookSink,err: %v", err)
	}

	assert.True(t, webhookSinkFilter == w.endpoint, "endpoint should be the same")
	assert.True(t, w.MockSend(newWebhookFilterEventPod), "newWebhookFilterEventPod should be matched.")
	assert.False(t, w.MockSend(kubeSystemNamespaceEventPod), "kubeSystemNamespaceEventPod should not be matched.")
}

func (ws *WebHookSink) MockSend(event *v1.Event) (matched bool) {
	for _, v := range ws.filters {
		if !v.Filter(event) {
			// this event should not be send
			return false
		}
	}
	// this event should be send, this return value just for testing
	return true
}
