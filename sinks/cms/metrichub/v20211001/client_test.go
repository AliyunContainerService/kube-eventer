package metrichub

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient_AppendHeader(t *testing.T) {

	c := &Client{
		stsToken: "not null",
	}
	request, err := http.NewRequest(http.MethodPost, "/", nil)
	assert.NoError(t, err)
	c.appendHeader(request)
	assert.NotNil(t, request.Header[XCmsCallerType])
	assert.NotNil(t, request.Header[XCmsSecurityToken])
}

// func makePanic(t *testing.T) func(error) {
// 	return func(err error) {
// 		fmt.Println(err)
// 		assert.Nil(t, err)
// 	}
// }

// func TestClient_SendZero(t *testing.T) {
// 	defer PanicCall(makePanic(t))
//
// 	_, err := client.PutSystemEvent(nil)
// 	assert.NoError(t, err)
// }
//
// func TestCreateMetricHubClient(t *testing.T) {
// 	defer PanicCall(makePanic(t))
//
// 	cfg := conf.GetConfigForTest(func(m map[string]interface{}) {
// 		delete(m["cloudMonitor"].(map[string]interface{}), "metricHubEndPoint")
// 	})
// 	assert.NotNil(t, cfg)
//
// 	myClient := CreateMetricHubClient(cfg)
// 	assert.NotNil(t, myClient)
// 	assert.NotEmpty(t, myClient.endPoint)
// }

func TestGetEndPoint(t *testing.T) {
	assert.Equal(t, "河源", GetEndPoint("cn-heyuan").Name)
	assert.Equal(t, "cn-hangzhou", GetEndPoint("xxxx").RegionId)
}
