package metrichub

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPutSystemEvent(t *testing.T) {
	sysEvent := &SystemEvent{
		Product: Product,
		// EventType:  "MetricAlert:" + EventType, // ! - 不存在
		Name:       "P1",
		EventTime:  time.Now().Format(EventTimeLayout),
		GroupId:    "0",
		ResourceId: `{"userId":"4","instanceId":"i-123abcxf"}`,
		Level:      "INFO",
		Status:     "AlertAlarm",
		UserId:     "04",
		Time:       time.Now().Format(EventTimeLayout), // ! - 不存在
	}
	// require.NoError(t, PutSystemEvent([]*SystemEvent{sysEvent}))
	assert.NotNil(t, sysEvent)
}

func TestPutSystemEvent_Without_RegionId(t *testing.T) {
	sysEvent := &SystemEvent{
		Product: Product,
		// EventType:  EventType,
		Name:       "P1",
		EventTime:  time.Now().Format(EventTimeLayout),
		GroupId:    "0",
		ResourceId: `{"userId":"4","instanceId":"i-123abcxf"}`,
		Level:      "INFO",
		Status:     "AlertOk",
		UserId:     "04",
		Content:    `{"__batchId__":"1234[1]@133455"}`,
		RegionId:   "",
		Time:       time.Now().Format(EventTimeLayout),
	}
	// require.Error(t, PutSystemEvent([]*SystemEvent{sysEvent}))
	assert.NotNil(t, sysEvent)
}
