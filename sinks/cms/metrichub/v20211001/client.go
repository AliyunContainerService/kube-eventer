package metrichub

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	log "k8s.io/klog"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// DefaultMetricHubEndPoint 云监控系统事件缺省上报地址
const (
	urlPath = "/event/system/upload"

	BatchCount   = 100        // 单次最大上报100条
	maxBodyBytes = 512 * 1024 // 单次最大不超过512K(http body最大值)

	DefaultRegionId = "cn-hangzhou"
)

type EndPoint struct {
	Name      string
	RegionId  string
	EndPoints []string
}

var (
	timeout10s = http.Client{Timeout: 10 * time.Second}

	// 数据来源：https://ata.alibaba-inc.com/articles/97388
	knownRegion = map[string]EndPoint{
		"cn-hangzhou":               {Name: "华东 1 (杭州)", RegionId: "cn-hangzhou", EndPoints: []string{"http://metrichub-cn-hangzhou.aliyun.com"}}, // (杭州、新加坡都可以访问)
		"cn-zhangjiakou":            {Name: "华北 3（张家口）", RegionId: "cn-zhangjiakou", EndPoints: []string{"http://metrichub-cn-zhangjiakou.aliyun.com"}},
		"cn-shanghai":               {Name: "华东 2 (上海)", RegionId: "cn-shanghai", EndPoints: []string{"http://metrichub-cn-shanghai.aliyun.com"}},
		"cn-shanghai-finance-1":     {Name: "上海金融云", RegionId: "cn-shanghai-finance-1", EndPoints: []string{"http://metrichub-cn-shanghai-finance-1.aliyun.com"}},
		"cn-beijing":                {Name: "华北 2 (北京)", RegionId: "cn-beijing", EndPoints: []string{"http://metrichub-cn-beijing.aliyun.com"}},
		"cn-qingdao":                {Name: "华北 1 (青岛)", RegionId: "cn-qingdao", EndPoints: []string{"http://metrichub-cn-qingdao.aliyun.com"}},
		"cn-shenzhen":               {Name: "华南 1 (深圳)", RegionId: "cn-shenzhen", EndPoints: []string{"http://metrichub-cn-shenzhen.aliyun.com"}},
		"cn-north-2-gov-1":          {Name: "政务云", RegionId: "cn-north-2-gov-1", EndPoints: []string{"http://metrichub-cn-north-2-gov-1.aliyun.com"}},
		"cn-shenzhen-finance-1":     {Name: "深圳金融云", RegionId: "cn-shenzhen-finance-1", EndPoints: []string{"http://metrichub-cn-shenzhen-finance-1.aliyun.com"}},
		"cn-hongkong":               {Name: "香港", RegionId: "cn-hongkong", EndPoints: []string{"http://metrichub-cn-hongkong.aliyun.com"}},
		"cn-huhehaote":              {Name: "华北 5 （呼和浩特）", RegionId: "cn-huhehaote", EndPoints: []string{"http://metrichub-cn-huhehaote.aliyun.com"}},
		"me-east-1":                 {Name: "中东东部 1（迪拜）", RegionId: "me-east-1", EndPoints: []string{"http://metrichub-me-east-1.aliyun.com"}},
		"us-west-1":                 {Name: "美国西部 1（硅谷 ）", RegionId: "us-west-1", EndPoints: []string{"http://metrichub-us-west-1.aliyun.com"}},
		"us-east-1":                 {Name: "美国东部 1（弗吉尼亚）", RegionId: "us-east-1", EndPoints: []string{"http://metrichub-us-east-1.aliyun.com"}},
		"ap-northeast-1":            {Name: "亚太东北 1 （日本 ）", RegionId: "ap-northeast-1", EndPoints: []string{"http://metrichub-ap-northeast-1.aliyun.com"}},
		"eu-central-1":              {Name: "欧洲中部 1（法兰克福）", RegionId: "eu-central-1", EndPoints: []string{"http://metrichub-eu-central-1.aliyun.com"}},
		"ap-southeast-2":            {Name: "亚太东南 2（悉尼）", RegionId: "ap-southeast-2", EndPoints: []string{"http://metrichub-ap-southeast-2.aliyun.com"}},
		"ap-southeast-1":            {Name: "亚太东南 1（新加坡）", RegionId: "ap-southeast-1", EndPoints: []string{"http://metrichub-ap-southeast-1.aliyun.com"}},
		"ap-southeast-3":            {Name: "亚太东南3（吉隆坡）", RegionId: "ap-southeast-3", EndPoints: []string{"http://metrichub-ap-southeast-3.aliyun.com"}},
		"cn-heyuan":                 {Name: "河源", RegionId: "cn-heyuan", EndPoints: []string{"http://metrichub-cn-heyuan.aliyun.com"}},
		"ap-south-1":                {Name: "印度-孟买", RegionId: "ap-south-1", EndPoints: []string{"http://metrichub-ap-south-1.aliyun.com", "http://metrichub-ap-south-1.aliyuncs.com"}},
		"ap-southeast-5":            {Name: "印尼-雅加达", RegionId: "ap-southeast-5", EndPoints: []string{"http://metrichub-ap-southeast-5.aliyun.com"}},
		"cn-chengdu":                {Name: "成都", RegionId: "cn-chengdu", EndPoints: []string{"http://metrichub-cn-chengdu.aliyun.com", "http://metrichub-cn-chengdu.aliyuncs.com"}},
		"cn-chengdu-smarthosting-1": {Name: "成都poc", RegionId: "cn-chengdu-smarthosting-1", EndPoints: []string{"http://metrichub-cn-chengdu-smarthosting-1.aliyun.com"}},
		"cn-zhengzhou-nebula-1":     {Name: "河南星云", RegionId: "cn-zhengzhou-nebula-1", EndPoints: []string{"http://metrichub-cn-zhengzhou-nebula-1.aliyun.com"}},
		"cn-wulanchabu":             {Name: "乌兰察布", RegionId: "cn-wulanchabu", EndPoints: []string{"http://metrichub-cn-wulanchabu.aliyun.com"}},
		"rus-west-1":                {Name: "俄罗斯(莫斯科)", RegionId: "rus-west-1", EndPoints: []string{"http://metrichub-rus-west-1.aliyun.com"}},
		"cn-huhehaote-nebula-1":     {Name: "内蒙古星云", RegionId: "cn-huhehaote-nebula-1", EndPoints: []string{"http://metrichub-cn-huhehaote-nebula-1.aliyun.com"}},
		"eu-west-1":                 {Name: "英国-伦敦", RegionId: "eu-west-1", EndPoints: []string{"http://metrichub-eu-west-1.aliyun.com", "http://metrichub-inner.eu-west-1.aliyuncs.com"}},
		"cn-guangzhou":              {Name: "广州(华南3)", RegionId: "cn-guangzhou", EndPoints: []string{"http://metrichub-cn-guangzhou.aliyun.com"}},
		"cn-zhangjiakou-spe":        {Name: "张北spe", RegionId: "cn-zhangjiakou-spe", EndPoints: []string{"http://metrichub-cn-zhangjiakou-spe.aliyun.com"}},
		"ap-southeast-6":            {Name: "菲律宾", RegionId: "ap-southeast-6", EndPoints: []string{"http://metrichub-ap-southeast-6.aliyun.com"}},
		"cn-shanghai-mybk":          {Name: "网商云", RegionId: "cn-shanghai-mybk", EndPoints: []string{"http://metrichub-cn-shanghai-mybk.aliyun.com"}},
		"cn-beijing-finance-1":      {Name: "北京金融云", RegionId: "cn-beijing-finance-1", EndPoints: []string{"http://metrichub-cn-beijing-finance-1.aliyuncs.com"}},
		"cn-nanjing":                {Name: "南京(华南5)", RegionId: "cn-nanjing", EndPoints: []string{"http://metrichub-cn-nanjing.aliyuncs.com"}},
		"ap-hochiminh-ant":          {Name: "越南(胡志明)蚂蚁", RegionId: "ap-hochiminh-ant", EndPoints: []string{"http://metrichub-ap-hochiminh-ant.aliyuncs.com"}},
		"ap-northeast-2":            {Name: "韩国(首尔) 3.5.6以上版本", RegionId: "ap-northeast-2", EndPoints: []string{"http://metrichub-ap-northeast-2.aliyuncs.com"}},
		"ap-southeast-7":            {Name: "泰国(曼谷) 3.5.6以上版本", RegionId: "ap-southeast-7", EndPoints: []string{"http://metrichub-ap-southeast-7.aliyuncs.com"}},
	}
)

func GetEndPoint(regionId string) EndPoint {
	v, ok := knownRegion[regionId]
	if !ok {
		v = knownRegion[DefaultRegionId]
	}
	return v
}

type Client struct {
	endPoint     string
	accessKeyId  string
	accessSecret string
	sourceIp     string
	stsToken     string
}

func HostIP() (r string) {
	addrSlice, err := net.InterfaceAddrs()
	if err == nil {
		ips := make([]net.IP, 0, len(addrSlice))
		for _, addr := range addrSlice {
			if ip, ok := addr.(*net.IPNet); ok && ip.IP.IsGlobalUnicast() {
				ips = append(ips, ip.IP)
			}
		}
		ipStr, _ := json.Marshal(ips)
		log.Info("ips: ", string(ipStr))
		if len(ips) > 0 {
			r = ips[0].String()
		}
	}
	return
}

func newClient(endPoint string, accessKeyId, accessSecret, stsToken string) *Client {
	r := &Client{
		endPoint:     strings.TrimSuffix(endPoint, "/"),
		accessKeyId:  accessKeyId,
		accessSecret: accessSecret,
		stsToken:     stsToken,
		sourceIp:     HostIP(),
	}

	return r
}

func CreateMetricHubClient(endPoint, accessKeyId, accessSecret, stsToken string) (r *Client) {
	log.Info("msg: ", "metric hub config", ", endPoint: ", endPoint, ", accessKeyId: ", accessKeyId)
	if endPoint == "" {
		endPoint = GetEndPoint("").EndPoints[0]
	}
	r = newClient(endPoint, accessKeyId, accessSecret, stsToken)
	return
}

func SafeClose(closer io.Closer) {
	if closer != nil {
		_ = closer.Close()
	}
}
func (p *Client) DoAction(request *http.Request, body []byte) (byte []byte, err error) {
	response, err := timeout10s.Do(request)
	if err == nil {
		defer SafeClose(response.Body)
		byte, _ = ioutil.ReadAll(response.Body)
	}
	errorLogIfNotNil(err, "uri", request.URL.String(), "headers", request.Header, "body", string(body))
	return
}

func (p *Client) appendHeader(request *http.Request) {
	request.Header[UserAgent] = []string{UserAgentValue}
	request.Header[Date] = []string{time.Now().Format(time.RFC1123)}
	request.Header[XCmsApiVersion] = []string{XCmsApiVersionValue}
	request.Header[XCmsSignature] = []string{XCmsSignatureMethod}
	request.Header[XCmsIp] = []string{p.sourceIp}

	if len(p.stsToken) > 0 {
		request.Header[XCmsCallerType] = []string{XCmsCallerToken}
		request.Header[XCmsSecurityToken] = []string{p.stsToken}
	}
}

type PutSystemEventResponse struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
}

func (p *PutSystemEventResponse) Success() bool {
	return "200" == p.Code
}

func ObjTypeName(obj interface{}, includePkg ...bool) (r string) {
	if obj != nil {
		typ := reflect.Indirect(reflect.ValueOf(obj)).Type()
		r = typ.Name()
		if r == "" || len(includePkg) == 0 || includePkg[0] {
			r = typ.String()
		}
	}
	return
}
func errorLogIfNotNil(err error, kvs ...interface{}) {
	for i := 0; i < len(kvs); i += 2 {
		if key, _ := kvs[i].(string); key != "" {
			sep := ", "
			if i == 0 {
				sep = ""
			}
			kvs[i] = sep + key + ": "
		}
	}
	kvs = append(kvs, ", errType: ", ObjTypeName(err), ", error: ", fmt.Sprintf("%+v", err))
	log.Error(kvs...)
}

// ErrExceed 大小超限
type ErrExceed struct {
	Size    int
	MaxSize int
}

func (e ErrExceed) Error() string {
	return fmt.Sprintf("size exceeds limit, max allowed is %v, actual %v", e.MaxSize, e.Size)
}

func (p *Client) PutSystemEvent(events []*SystemEvent) (response PutSystemEventResponse, err error) {
	if len(events) <= 0 {
		return
	}

	jsonBytes, _ := json.Marshal(events)
	if len(jsonBytes) > maxBodyBytes {
		err = ErrExceed{MaxSize: maxBodyBytes, Size: len(jsonBytes)}
		return
	}
	request, _ := http.NewRequest(http.MethodPost, p.endPoint+urlPath, bytes.NewReader(jsonBytes))
	request.Header[ContentMd5] = []string{fmt.Sprintf("%X", md5.Sum(jsonBytes))}
	request.Header[ContentType] = []string{ContentJson} // must be application/json，without utf-8
	p.appendHeader(request)

	// 处理签名
	var digest string
	if digest, _, err = p.SignatureRequest(request); err == nil {
		request.Header[Authorization] = []string{fmt.Sprintf("%v:%v", p.accessKeyId, digest)}

		// 发送请求
		var respJsonBytes []byte
		respJsonBytes, err = p.DoAction(request, jsonBytes)
		if err == nil {
			if err = json.Unmarshal(respJsonBytes, &response); err == nil && !response.Success() {
				err = errors.New(response.Message)
			}
		}
	}
	errorLogIfNotNil(err, "uri", request.URL.String(), "headers", request.Header, "body", string(jsonBytes))
	return response, err
}
