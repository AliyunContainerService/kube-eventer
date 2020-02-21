package dingtalk

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TEST_CLUSTERID        = "abcdefghijklmnopqrstuvwxyz"
	TEST_REGION           = "cn-shenzhen"
	TEST_NODENAME         = "cn-shenzhen.i-xxxxxxxxx"
	TEST_DEPLOY_NAME      = "testdeploy"
	TEST_POD_NAME         = "testpod"
	TEST_STATEFULSET_NAME = "testss"
	TEST_DAEMONSET_NAME   = "testds"
	TEST_SERVICE_NAME     = "testservice"
	TEST_CROBJOB          = "logs-cleaner"
	TEST_NAMESPACE        = "default"
	TEST_RESOURCE_TYPE    = "Deployment"
)

func TestNewMarkdownMsgBuilder_Deployment(t *testing.T) {
	e := createTestEvent()
	e.Name = TEST_DEPLOY_NAME
	e.InvolvedObject.Kind = "Deployment"
	e.Namespace = TEST_NAMESPACE
	m := NewMarkdownMsgBuilder(TEST_CLUSTERID, TEST_REGION, e)
	text := m.Build()
	t.Log(string(text))
	assert.True(t, m != nil)
}

func TestNewMarkdownMsgBuilder_Pod(t *testing.T) {
	e := createTestEvent()
	e.Name = TEST_POD_NAME
	e.InvolvedObject.Kind = "Pod"
	e.Namespace = TEST_NAMESPACE
	m := NewMarkdownMsgBuilder(TEST_CLUSTERID, TEST_REGION, e)
	m.AddLabels(createTestLabels())
	text := m.Build()
	t.Log(string(text))
	assert.True(t, m != nil)
}

func TestNewMarkdownMsgBuilder_StatefulSet(t *testing.T) {
	e := createTestEvent()
	e.Name = TEST_STATEFULSET_NAME
	e.InvolvedObject.Kind = "StatefulSet"
	e.Namespace = TEST_NAMESPACE
	m := NewMarkdownMsgBuilder(TEST_CLUSTERID, TEST_REGION, e)
	m.AddLabels(createTestLabels())
	text := m.Build()
	t.Log(string(text))
	assert.True(t, m != nil)
}

func TestNewMarkdownMsgBuilder_DaemonSet(t *testing.T) {
	e := createTestEvent()
	e.Name = TEST_DAEMONSET_NAME
	e.InvolvedObject.Kind = "DaemonSet"
	e.Namespace = "kube-system"
	m := NewMarkdownMsgBuilder(TEST_CLUSTERID, TEST_REGION, e)
	m.AddLabels(createTestLabels())
	text := m.Build()
	t.Log(string(text))
	assert.True(t, m != nil)
}

func TestNewMarkdownMsgBuilder_CronJob(t *testing.T) {
	e := createTestEvent()
	e.Name = TEST_CROBJOB
	e.InvolvedObject.Kind = "CronJob"
	e.Namespace = TEST_NAMESPACE
	m := NewMarkdownMsgBuilder(TEST_CLUSTERID, TEST_REGION, e)
	m.AddLabels(createTestLabels())
	text := m.Build()
	t.Log(string(text))
	assert.True(t, m != nil)
}

func TestNewMarkdownMsgBuilder_Service(t *testing.T) {
	e := createTestEvent()
	e.Name = TEST_SERVICE_NAME
	e.InvolvedObject.Kind = "Service"
	e.Namespace = TEST_NAMESPACE
	m := NewMarkdownMsgBuilder(TEST_CLUSTERID, TEST_REGION, e)
	m.AddLabels(createTestLabels())
	text := m.Build()
	t.Log(string(text))
	assert.True(t, m != nil)
}

func TestRemoveDotContent(t *testing.T) {
	s := removeDotContent("eventer.15b21c773eb1181a.sssss")
	t.Log(s)
	assert.True(t, !strings.ContainsAny(s, "."))
}

func TestAddNodeName(t *testing.T) {
	e := createTestEvent()
	e.Name = TEST_SERVICE_NAME
	e.InvolvedObject.Kind = "Service"
	e.Namespace = TEST_NAMESPACE
	m := NewMarkdownMsgBuilder(TEST_CLUSTERID, TEST_REGION, e)
	m.AddNodeName(TEST_NODENAME)
	t.Log(m.OutputText)
}

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

func createTestLabels() []string {
	return []string{"a", "b"}

}
