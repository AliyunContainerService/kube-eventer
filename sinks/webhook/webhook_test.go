package webhook

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"testing"
)

const (
	webhookSinkValid = "https://oapi.dingtalk.com/robot/send?access_token=token&level=Normal&namespaces=a,b&kinds=c,d&header=contentType=demo&header=content2=3"
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
