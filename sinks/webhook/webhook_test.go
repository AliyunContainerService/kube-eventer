package webhook

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	webhookSink            = "https://oapi.dingtalk.com/robot/send?access_token=token&level=Warning&namespaces=kube-system&kinds=Pod&header=contentType=demo&header=content2=3"
	webhookSinkReasonRegex = "https://oapi.dingtalk.com/robot/send?access_token=token&level=Warning&namespaces=kube-system&reason=[^Failed*]&kinds=Pod&header=contentType=demo&header=content2=3"
)

var (
	newEvent = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "Event1",
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Namespace: "kube-system",
		},
		Reason:  "FailedStartUp",
		Type:    Warning,
		Message: "DEMO",
	}
)

func TestNewWebhookSinkReasonRegex(t *testing.T) {
	uri, err := url.Parse(webhookSinkReasonRegex)
	if err != nil {
		t.Fatalf("Failed to prase webhookSinkValid,err: %v", err)
	}
	w, err := NewWebHookSink(uri)
	if err != nil {
		t.Fatalf("Failed to create NewWebhookSink,err: %v", err)
	}

	assert.True(t, w.MockSend(newEvent), "newEvent should be matched.")
}

func TestNewWebhookSink(t *testing.T) {
	uri, err := url.Parse(webhookSink)
	if err != nil {
		t.Fatalf("Failed to prase webhookSinkValid,err: %v", err)
	}
	w, err := NewWebHookSink(uri)
	if err != nil {
		t.Fatalf("Failed to create NewWebhookSink,err: %v", err)
	}

	assert.True(t, webhookSink == w.endpoint, "endpoint should be the same")
}

func TestNewWebhookFilter(t *testing.T) {
	uri, err := url.Parse(webhookSink)
	if err != nil {
		t.Fatalf("Failed to prase webhookSinkFilter,err: %v", err)
	}
	w, err := NewWebHookSink(uri)
	if err != nil {
		t.Fatalf("Failed to create NewWebhookSink,err: %v", err)
	}

	assert.True(t, w.MockSend(newEvent), "newEvent should be matched.")
}

func TestRenderMessageWithDoubleQuote(t *testing.T) {
	uri, err := url.Parse(webhookSink)
	if err != nil {
		t.Fatalf("Failed to prase webhookSinkFilter,err: %v", err)
	}
	w, err := NewWebHookSink(uri)
	if err != nil {
		t.Fatalf("Failed to create NewWebhookSink,err: %v", err)
	}
	event := &v1.Event{
		Type:    Warning,
		Message: "pod \"demo-1rare3\" OOMKilled",
	}
	w.bodyTemplate = `{"EventMessage": "{{ .Message }}"}`
	template, _ := w.RenderBodyTemplate(event)
	assert.Equal(t, `{"EventMessage": "pod demo-1rare3 OOMKilled"}`, template)
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
