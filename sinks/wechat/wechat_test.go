package wechat

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"testing"
	"time"
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
	u, _ := url.Parse("wechat:?corp_id=xxxxxxxx&corp_secret=xxxxxxxxx&agent_id=111111&to_user=<user>&label=<label>&level=Normal")
	d, _ := NewWechatSink(u)
	msg := createMsgFromEvent(d, event)
	text, _ := json.Marshal(msg)
	t.Log(string(text))
	// t.Log(msg.Text)
	assert.True(t, msg != nil)
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
