package feishu

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"testing"
	"time"
)

const (
	TEST_CLUSTERID        = "abcdefghijklmnopqrstuvwxyz"
	TEST_REGION           = "cn-shenzhen"
	TEST_NODENAME         = "lcc-system"
	TEST_DEPLOY_NAME      = "testdeploy"
	TEST_POD_NAME         = "testpod"
	TEST_STATEFULSET_NAME = "testss"
	TEST_DAEMONSET_NAME   = "testds"
	TEST_SERVICE_NAME     = "testservice"
	TEST_CROBJOB          = "logs-cleaner"
	TEST_NAMESPACE        = "default"
	TEST_RESOURCE_TYPE    = "Deployment"
)

func TestGetLevel(t *testing.T) {
	warning := getLevel(v1.EventTypeWarning)
	normal := getLevel(v1.EventTypeNormal)
	none := getLevel("")
	assert.True(t, warning > normal)
	assert.True(t, warning == WARNING)
	assert.True(t, normal == NORMAL)
	assert.True(t, 0 == none)
}

func TestCreateMsgFromEvent(t *testing.T) {
	labels := make([]string, 2)
	labels[0] = "abcd"
	labels[1] = "defg"
	event := createTestEvent()
	event.Source.Host = TEST_NODENAME
	event.InvolvedObject.Kind = TEST_RESOURCE_TYPE
	event.Name = TEST_DEPLOY_NAME
	event.Namespace = TEST_NAMESPACE
	u, _ := url.Parse("feishu:https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxxxx?label=<label>&level=Normal&cluster_id=cloud-prod")
	f, _ := NewFeishuSink(u)
	//	f.Labels = labels
	f.Send(event)
}

/*
type FeishuSink struct {
	BotToken   string
	Namespaces []string
	Kinds      []string
	Level      int
	Labels     []string
	ClusterID  string
	Region     string
}
*/

func createTestEvent() *v1.Event {
	now := time.Now()
	event := &v1.Event{
		Reason:         "996 work schedule",
		Message:        "on the way to icu",
		Count:          251,
		LastTimestamp:  metav1.NewTime(now),
		FirstTimestamp: metav1.NewTime(now),
		Type:           "Warning",
	}
	return event
}
