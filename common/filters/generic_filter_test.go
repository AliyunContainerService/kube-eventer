package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	emptyNamespaceEventPod = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "Event0",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
		Reason: "RouteFailedToBeCreated",
	}

	defaultNamespaceEventPod = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Event1",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Reason: "CreateInitContainerFailed",
	}

	kubeSystemNamespaceEventPod = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Event2",
			Namespace: "kube-system",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Reason: "FailedStartUp",
	}

	SuccessfulReasonEvent = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Event2",
			Namespace: "kube-system",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Reason: "SuccessfulReason",
		Type:   v1.EventTypeNormal,
	}

	emptyReasonEventPod = &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Event3",
			Namespace: "kube-system",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
	}
)

func TestKindFilter(t *testing.T) {
	kindFilter := NewGenericFilter("Kind", []string{"Node"}, false)

	assert.True(t, kindFilter.Filter(emptyNamespaceEventPod), "Node's event should not be matched.")
	assert.False(t, kindFilter.Filter(defaultNamespaceEventPod), "Pod's event should be matched.")
	assert.False(t, kindFilter.Filter(kubeSystemNamespaceEventPod), "Pod's event should be matched.")
}

func TestNamespaceFilter(t *testing.T) {
	namespaceFilter := NewGenericFilter("Namespace", []string{"default", "kube-system"}, false)

	assert.False(t, namespaceFilter.Filter(emptyNamespaceEventPod), "EmptyNamespace should not be matched.")
	assert.True(t, namespaceFilter.Filter(defaultNamespaceEventPod), "default namespace should be matched.")
	assert.True(t, namespaceFilter.Filter(kubeSystemNamespaceEventPod), "kube-system namespace should be matched.")
}

func TestReasonFilter(t *testing.T) {
	reasonFilter := NewGenericFilter("Reason", []string{"Failed"}, true)

	assert.True(t, reasonFilter.Filter(emptyNamespaceEventPod), "FailedReason should not be matched.")
	assert.True(t, reasonFilter.Filter(defaultNamespaceEventPod), "FailedReason should be matched.")
	assert.True(t, reasonFilter.Filter(kubeSystemNamespaceEventPod), "FailedReason should be matched.")
	assert.False(t, reasonFilter.Filter(SuccessfulReasonEvent), "SuccessfulReason should be not matched.")
	assert.False(t, reasonFilter.Filter(emptyReasonEventPod), "EmptyReason should be not matched.")
}

func TestComplexReasonFilter(t *testing.T) {
	reasonFilter := NewGenericFilter("Reason", []string{"(Failed|Success)"}, true)

	assert.True(t, reasonFilter.Filter(emptyNamespaceEventPod), "FailedReason should not be matched.")
	assert.True(t, reasonFilter.Filter(defaultNamespaceEventPod), "FailedReason should be matched.")
	assert.True(t, reasonFilter.Filter(kubeSystemNamespaceEventPod), "FailedReason should be matched.")
	assert.True(t, reasonFilter.Filter(SuccessfulReasonEvent), "SuccessfulReason should be matched.")
	assert.False(t, reasonFilter.Filter(emptyReasonEventPod), "EmptyReason should be not matched.")
}

func TestEmptyReasonFilter(t *testing.T) {
	// when enable regexp, empty regexp key should be return true
	emptyReasonFilter := NewGenericFilter("Reason", []string{}, true)

	assert.True(t, emptyReasonFilter.Filter(emptyNamespaceEventPod), "FailedReason should be matched.")
	assert.True(t, emptyReasonFilter.Filter(defaultNamespaceEventPod), "FailedReason should be matched.")
	assert.True(t, emptyReasonFilter.Filter(kubeSystemNamespaceEventPod), "FailedReason should be matched.")
	assert.True(t, emptyReasonFilter.Filter(SuccessfulReasonEvent), "SuccessfulReason should be matched.")
	assert.True(t, emptyReasonFilter.Filter(emptyReasonEventPod), "EmptyReason should be matched.")
}
