package dingtalk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/olekukonko/tablewriter"
	//"os"
	"encoding/json"
	"net/url"
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
	u, _ := url.Parse("dingtalk:https://oapi.dingtalk.com/robot/send?access_token=<access_token>&label=<label>&level=Normal")
	d, _ := NewDingTalkSink(u)
	d.Labels = labels
	msg := createMsgFromEvent(d, event)
	text, _ := json.Marshal(msg)
	t.Log("\n" + string(text))
	// t.Log(msg.Text)
	assert.True(t, msg != nil)
}

func TestCreateMsgFromEvent_Markdown(t *testing.T) {
	labels := make([]string, 2)
	labels[0] = "abcd"
	labels[1] = "defg"
	event := createTestEvent()
	event.Source.Host = TEST_NODENAME
	event.InvolvedObject.Kind = TEST_RESOURCE_TYPE
	event.Name = TEST_DEPLOY_NAME
	event.Namespace = TEST_NAMESPACE
	u, _ := url.Parse("dingtalk:https://oapi.dingtalk.com/robot/send?access_token=<access_token>&label=<label>&level=Normal" + "&msg_type=markdown&cluster_id=" + TEST_CLUSTERID + "&region=" + TEST_REGION)
	d, _ := NewDingTalkSink(u)
	d.Labels = labels
	msg := createMsgFromEvent(d, event)
	text, _ := json.Marshal(msg)
	t.Log("\n" + string(text))
	// t.Log(msg.Text)
	assert.True(t, msg != nil)
}
