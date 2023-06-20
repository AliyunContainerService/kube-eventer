package sls

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"
	"testing"
)

const (
	Warning = "Warning"
	Normal  = "Normal"
)

func TestSLSSinkParse(t *testing.T) {
	u, _ := url.Parse("sls:https://sls.aliyuncs.com?internal=true&logStore=k8s-event&project=test_projectId&topic=&label=ClusterId,test_clusterId&label=RegionId,test_regionId&label=UserId,test_uid")
	d, _ := NewSLSSink(u)
	t.Logf("sls sink config: %v", d.Config)
}

func TestSLSEventToContents(t *testing.T) {
	u, _ := url.Parse("sls:https://sls.aliyuncs.com?internal=true&logStore=k8s-event&project=test_projectId&topic=&label=ClusterId,test_clusterId&label=RegionId,test_regionId&label=UserId,test_uid&label=ClusterName,abasdfasdf")
	d, _ := NewSLSSink(u)

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

	d.eventToContents(newEvent)

	t.Logf("sls sink config: %v", d.Config)
}
