package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

var (
	TestEvent = &v1.Event{
		Type: "Warning",
		InvolvedObject: v1.ObjectReference{
			Kind:      "Node",
			Namespace: "default",
		},
		Reason: "BackOff",
	}
)

func TestEvents(t *testing.T) {
	kindFilter := NewGenericFilter("Kind", []string{"Node"}, false)
	assert.True(t, kindFilter.Filter(TestEvent), "")

	typeFilter := NewGenericFilter("Type", []string{"Warning"}, false)
	assert.True(t, typeFilter.Filter(TestEvent), "")

	namespaceFilter := NewGenericFilter("Namespace", []string{"default"}, false)
	assert.True(t, namespaceFilter.Filter(TestEvent), "")

	reasonFilter := NewGenericFilter("Reason", []string{"BackOff"}, false)
	assert.True(t, reasonFilter.Filter(TestEvent), "")

	regexReasonFilter := NewGenericFilter("Reason", []string{"BackOff"}, true)
	assert.True(t, regexReasonFilter.Filter(TestEvent), "")

	regexReasonsFilter := NewGenericFilter("Reason", []string{"Unhealthy", "BackOff"}, true)
	assert.True(t, regexReasonsFilter.Filter(TestEvent), "")
}
