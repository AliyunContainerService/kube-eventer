package sls

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	Warning = "Warning"
	Normal  = "Normal"
)

func TestParseConfig(t *testing.T) {
	u, err := url.Parse("sls:https://sls.aliyuncs.com?internal=true&logStore=k8s-event&project=test_projectId&topic=&label=ClusterId,test_clusterId&label=RegionId,test_regionId&label=UserId,test_uid")
	assert.NoError(t, err, "parse url")
	cfg, err := parseConfig(u)
	assert.NoError(t, err, "parse config")
	t.Logf("sls sink config: %v", cfg)
}

func TestParseLabels(t *testing.T) {
	testCases := []struct {
		name     string
		labels   []string
		expected map[string]string
	}{
		{
			name:     "labels is empty",
			labels:   []string{},
			expected: map[string]string{},
		},
		{
			name:     "invalid labels",
			labels:   []string{"key,value,other"},
			expected: map[string]string{},
		},
		{
			name:     "valid labels",
			labels:   []string{"key,value"},
			expected: map[string]string{"key": "value"},
		},
		{
			name:     "valid and invalid labels",
			labels:   []string{"key1,value1", "key2,value2,other"},
			expected: map[string]string{"key1": "value1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := parseLabels(tc.labels)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestSLSEventToContents(t *testing.T) {
	newEvent := &v1.Event{
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

	contents := eventToContents(newEvent, nil)
	t.Logf("contents: %v", contents)
}

func TestGetSLSEndpoint(t *testing.T) {
	testCases := []struct {
		name     string
		region   string
		internal bool
		expected string
	}{
		{
			name:     "default value",
			region:   "",
			internal: false,
			expected: SLSDefaultEndpoint,
		},
		{
			name:     "external",
			region:   "",
			internal: false,
			expected: SLSDefaultEndpoint,
		},
		{
			name:     "internal",
			region:   "cn-beijing",
			internal: true,
			expected: "cn-beijing-intranet.log.aliyuncs.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := getSLSEndpoint(tc.region, tc.internal)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
