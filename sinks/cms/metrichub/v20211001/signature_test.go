package metrichub

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestSignature_OK(t *testing.T) {
	// 该用例参考: https://ata.alibaba-inc.com/articles/87434
	p := &Client{
		accessSecret: "testsecret",
	}
	request, err := http.NewRequest(http.MethodPost, "http://metrichub.aliyun-inc.com/metric/custom/upload", nil)
	require.NoError(t, err)
	request.Header[ContentMd5] = []string{"0B9BE351E56C90FED853B32524253E8B"}
	request.Header[ContentType] = []string{"application/json"}
	request.Header[Date] = []string{"Tue, 11 Dec 2018 21:05:51 +0800"}
	request.Header[XCmsApiVersion] = []string{"1.0"}
	request.Header[XCmsIp] = []string{"127.0.0.1"}
	request.Header[XCmsSignature] = []string{XCmsSignatureMethod}
	digest, signStr, err := p.SignatureRequest(request)
	require.NoError(t, err)
	fmt.Println(signStr)
	const expectSignStr = `POST
0B9BE351E56C90FED853B32524253E8B
application/json
Tue, 11 Dec 2018 21:05:51 +0800
x-cms-api-version:1.0
x-cms-ip:127.0.0.1
x-cms-signature:hmac-sha1
/metric/custom/upload`
	require.Equal(t, expectSignStr, signStr)
	fmt.Println("Signature: " + digest)
	require.Equal(t, "1DC19ED63F755ACDE203614C8A1157EB1097E922", digest)
}

func TestSignature_WithoutDate(t *testing.T) {
	p := &Client{
		accessSecret: "testsecret",
	}
	digest, signStr, err := p.Signature(http.MethodPost, "/metric/custom/upload", http.Header{
		ContentMd5:     {"0B9BE351E56C90FED853B32524253E8B"},
		ContentType:    {"application/json"},
		XCmsApiVersion: {"1.0"},
		XCmsIp:         {"127.0.0.1"},
		XCmsSignature:  {XCmsSignatureMethod},
		//Date:           "Tue, 11 Dec 2018 21:05:51 +0800",
	})
	require.Error(t, err)
	assert.Empty(t, digest)
	assert.Empty(t, signStr)
}

func TestSignature_QueryParams(t *testing.T) {
	p := &Client{
		accessSecret: "testsecret",
	}
	digest, signStr, err := p.Signature(http.MethodPost, "/metric/custom/upload?name=hcj&employeeId=11", http.Header{
		ContentMd5:     {"0B9BE351E56C90FED853B32524253E8B"},
		ContentType:    {"application/json"},
		Date:           {"Tue, 11 Dec 2018 21:05:51 +0800"},
		XCmsApiVersion: {"1.0"},
		XCmsIp:         {"127.0.0.1"},
		XCmsSignature:  {XCmsSignatureMethod},
	})
	require.NoError(t, err)
	fmt.Println(signStr)
	const expectSignStr = `POST
0B9BE351E56C90FED853B32524253E8B
application/json
Tue, 11 Dec 2018 21:05:51 +0800
x-cms-api-version:1.0
x-cms-ip:127.0.0.1
x-cms-signature:hmac-sha1
/metric/custom/upload?employeeId=11&name=hcj`
	assert.Equal(t, expectSignStr, signStr)
	fmt.Println("Signature: " + digest)
	assert.Equal(t, "D9F6DDC52C035D7F28418226DA91B3E9C4380556", digest)
}
