package sls

import (
	"net/url"
	"testing"
)

func TestSLSSinkParse(t *testing.T) {
	u, _ := url.Parse("sls:https://sls.aliyuncs.com?internal=true&logStore=k8s-event&project=test_projectId&topic=&label=ClusterId,test_clusterId&label=RegionId,test_regionId&label=UserId,test_uid")
	d, _ := NewSLSSink(u)
	t.Logf("sls sink config: %v", d.Config)
}
