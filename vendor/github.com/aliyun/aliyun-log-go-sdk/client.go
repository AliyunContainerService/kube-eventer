package sls

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/aliyun-log-go-sdk/util"
)

// GlobalForceUsingHTTP if GlobalForceUsingHTTP is true, then all request will use HTTP(ignore LogProject's UsingHTTP flag)
var GlobalForceUsingHTTP = false

// RetryOnServerErrorEnabled if RetryOnServerErrorEnabled is false, then all error requests will not be retried
var RetryOnServerErrorEnabled = true

var GlobalDebugLevel = 0

var MaxCompletedRetryCount = 20

var MaxCompletedRetryLatency = 5 * time.Minute

// compress type
const (
	Compress_LZ4  = iota // 0
	Compress_None        // 1
	Compress_ZSTD        // 2
	Compress_Max         // max compress type(just for filter invalid compress type)
)

var InvalidCompressError = errors.New("Invalid Compress Type")

const DefaultLogUserAgent = "golang-sdk-v0.1.0"

// AuthVersionType the version of auth
type AuthVersionType string

const (
	AuthV0 AuthVersionType = "v0"
	// AuthV1 v1
	AuthV1 AuthVersionType = "v1"
	// AuthV4 v4
	AuthV4 AuthVersionType = "v4"
)

// Error defines sls error
type Error struct {
	HTTPCode  int32  `json:"httpCode"`
	Code      string `json:"errorCode"`
	Message   string `json:"errorMessage"`
	RequestID string `json:"requestID"`
}

func IsDebugLevelMatched(level int) bool {
	return level <= GlobalDebugLevel
}

// NewClientError new client error
func NewClientError(err error) *Error {
	if err == nil {
		return nil
	}
	if clientError, ok := err.(*Error); ok {
		return clientError
	}
	clientError := new(Error)
	clientError.HTTPCode = -1
	clientError.Code = "ClientError"
	clientError.Message = err.Error()
	return clientError
}

func (e Error) String() string {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

func (e Error) Error() string {
	return e.String()
}

func IsTokenError(err error) bool {
	if clientErr, ok := err.(*Error); ok {
		if clientErr.HTTPCode == 401 {
			return true
		}
	}
	return false
}

// Client ...
type Client struct {
	Endpoint        string // IP or hostname of SLS endpoint
	AccessKeyID     string // Deprecated: use credentialsProvider instead
	AccessKeySecret string // Deprecated: use credentialsProvider instead
	SecurityToken   string // Deprecated: use credentialsProvider instead
	UserAgent       string // default defaultLogUserAgent
	RequestTimeOut  time.Duration
	RetryTimeOut    time.Duration
	HTTPClient      *http.Client
	Region          string
	AuthVersion     AuthVersionType //  v1 or v4 signature,default is v1

	accessKeyLock       sync.RWMutex
	credentialsProvider CredentialsProvider
	// User defined common headers.
	// When conflict with sdk pre-defined headers, the value will
	// be ignored
	CommonHeaders map[string]string
	InnerHeaders  map[string]string
}

func convert(c *Client, projName string) *LogProject {
	c.accessKeyLock.RLock()
	defer c.accessKeyLock.RUnlock()
	return convertLocked(c, projName)
}

func convertLocked(c *Client, projName string) *LogProject {
	var p *LogProject
	if c.credentialsProvider != nil {
		p, _ = NewLogProjectV2(projName, c.Endpoint, c.credentialsProvider)
	} else { // back compatible
		p, _ = NewLogProject(projName, c.Endpoint, c.AccessKeyID, c.AccessKeySecret)
	}

	p.SecurityToken = c.SecurityToken
	p.UserAgent = c.UserAgent
	p.AuthVersion = c.AuthVersion
	p.Region = c.Region
	p.CommonHeaders = c.CommonHeaders
	p.InnerHeaders = c.InnerHeaders
	if c.HTTPClient != nil {
		p.httpClient = c.HTTPClient
	}
	if c.RequestTimeOut != time.Duration(0) {
		p.WithRequestTimeout(c.RequestTimeOut)
	}
	if c.RetryTimeOut != time.Duration(0) {
		p.WithRetryTimeout(c.RetryTimeOut)
	}

	return p
}

// Set credentialsProvider for client and returns the same client.
func (c *Client) WithCredentialsProvider(provider CredentialsProvider) *Client {
	c.credentialsProvider = provider
	return c
}

// SetUserAgent set a custom userAgent
func (c *Client) SetUserAgent(userAgent string) {
	c.UserAgent = userAgent
}

// SetHTTPClient set a custom http client, all request will send to sls by this client
func (c *Client) SetHTTPClient(client *http.Client) {
	c.HTTPClient = client
}

// SetRetryTimeout set retry timeout
func (c *Client) SetRetryTimeout(timeout time.Duration) {
	c.RetryTimeOut = timeout
}

// SetAuthVersion set signature version that the client used
func (c *Client) SetAuthVersion(version AuthVersionType) {
	c.accessKeyLock.Lock()
	c.AuthVersion = version
	c.accessKeyLock.Unlock()
}

// SetRegion set a region, must be set if using signature version v4
func (c *Client) SetRegion(region string) {
	c.accessKeyLock.Lock()
	c.Region = region
	c.accessKeyLock.Unlock()
}

// ResetAccessKeyToken reset client's access key token
func (c *Client) ResetAccessKeyToken(accessKeyID, accessKeySecret, securityToken string) {
	c.accessKeyLock.Lock()
	c.AccessKeyID = accessKeyID
	c.AccessKeySecret = accessKeySecret
	c.SecurityToken = securityToken
	c.credentialsProvider = NewStaticCredentialsProvider(accessKeyID, accessKeySecret, securityToken)
	c.accessKeyLock.Unlock()
}

// CreateProject create a new loghub project.
func (c *Client) CreateProject(name, description string) (*LogProject, error) {
	return c.CreateProjectV2(name, description, "")
}

// CreateProjectV2 create a new loghub project, with dataRedundancyType option.
func (c *Client) CreateProjectV2(name, description, dataRedundancyType string) (*LogProject, error) {
	type Body struct {
		ProjectName        string `json:"projectName"`
		Description        string `json:"description"`
		DataRedundancyType string `json:"dataRedundancyType,omitempty"`
	}
	body, err := json.Marshal(Body{
		ProjectName:        name,
		Description:        description,
		DataRedundancyType: dataRedundancyType,
	})
	if err != nil {
		return nil, err
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%d", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}

	uri := "/"
	proj := convert(c, name)
	resp, err := request(proj, "POST", uri, h, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return proj, nil
}

// UpdateProject create a new loghub project.
func (c *Client) UpdateProject(name, description string) (*LogProject, error) {
	type Body struct {
		Description string `json:"description"`
	}
	body, err := json.Marshal(Body{
		Description: description,
	})
	if err != nil {
		return nil, err
	}

	h := map[string]string{
		"x-log-bodyrawsize": fmt.Sprintf("%d", len(body)),
		"Content-Type":      "application/json",
		"Accept-Encoding":   "deflate", // TODO: support lz4
	}
	uri := "/"
	proj := convert(c, name)
	_, err = request(proj, "PUT", uri, h, body)
	if err != nil {
		return nil, err
	}

	return proj, nil
}

// GetProject ...
func (c *Client) GetProject(name string) (*LogProject, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	uri := "/"
	proj := convert(c, name)
	resp, err := request(proj, "GET", uri, h, nil)
	if err != nil {
		return nil, NewClientError(err)
	}
	defer resp.Body.Close()
	buf, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		err := new(Error)
		json.Unmarshal(buf, err)
		return nil, err
	}
	err = json.Unmarshal(buf, proj)
	return proj, err
}

// ListProject list all projects in specific region
// the region is related with the client's endpoint
func (c *Client) ListProject() (projectNames []string, err error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	uri := "/"
	proj := convert(c, "")

	type Project struct {
		ProjectName string `json:"projectName"`
	}

	type Body struct {
		Projects []Project `json:"projects"`
	}

	r, err := request(proj, "GET", uri, h, nil)
	if err != nil {
		return nil, NewClientError(err)
	}

	defer r.Body.Close()
	buf, _ := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		json.Unmarshal(buf, err)
		return nil, err
	}

	body := &Body{}
	err = json.Unmarshal(buf, body)
	for _, project := range body.Projects {
		projectNames = append(projectNames, project.ProjectName)
	}
	return projectNames, err
}

// ListProjectV2 list all projects in specific region
// the region is related with the client's endpoint
// ref https://www.alibabacloud.com/help/doc-detail/74955.htm
func (c *Client) ListProjectV2(offset, size int) (projects []LogProject, count, total int, err error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	urlVal := url.Values{}
	urlVal.Add("offset", strconv.Itoa(offset))
	urlVal.Add("size", strconv.Itoa(size))
	uri := fmt.Sprintf("/?%s", urlVal.Encode())
	proj := convert(c, "")

	type Body struct {
		Projects []LogProject `json:"projects"`
		Count    int          `json:"count"`
		Total    int          `json:"total"`
	}

	r, err := request(proj, "GET", uri, h, nil)
	if err != nil {
		return nil, 0, 0, NewClientError(err)
	}

	defer r.Body.Close()
	buf, _ := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK {
		err := new(Error)
		json.Unmarshal(buf, err)
		return nil, 0, 0, err
	}

	body := &Body{}
	err = json.Unmarshal(buf, body)
	return body.Projects, body.Count, body.Total, err
}

// CheckProjectExist check project exist or not
func (c *Client) CheckProjectExist(name string) (bool, error) {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}
	uri := "/"
	proj := convert(c, name)
	resp, err := request(proj, "GET", uri, h, nil)
	if err != nil {
		if _, ok := err.(*Error); ok {
			slsErr := err.(*Error)
			if slsErr.Code == "ProjectNotExist" {
				return false, nil
			}
			return false, slsErr
		}
		return false, err
	}
	defer resp.Body.Close()
	return true, nil
}

// DeleteProject ...
func (c *Client) DeleteProject(name string) error {
	h := map[string]string{
		"x-log-bodyrawsize": "0",
	}

	proj := convert(c, name)
	uri := "/"
	resp, err := request(proj, "DELETE", uri, h, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Close the client
func (c *Client) Close() error {
	return nil
}

func (c *Client) setSignV4IfInAcdr(endpoint string) {
	region, err := util.ParseRegion(endpoint)
	if err == nil && strings.Contains(region, "-acdr-ut-") {
		c.AuthVersion = AuthV4
		c.Region = region
	}
}
