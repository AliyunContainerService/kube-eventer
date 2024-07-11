package sls

// request sends a request to SLS.
import (
	"bytes"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/go-kit/kit/log/level"
)

// request sends a request to alibaba cloud Log Service.
// @note if error is nil, you must call http.Response.Body.Close() to finalize reader
func (c *Client) request(project, method, uri string, headers map[string]string, body []byte) (*http.Response, error) {
	// The caller should provide 'x-log-bodyrawsize' header
	if _, ok := headers[HTTPHeaderBodyRawSize]; !ok {
		return nil, fmt.Errorf("Can't find 'x-log-bodyrawsize' header")
	}

	var endpoint string
	var usingHTTPS bool
	if strings.HasPrefix(c.Endpoint, "https://") {
		endpoint = c.Endpoint[8:]
		usingHTTPS = true
	} else if strings.HasPrefix(c.Endpoint, "http://") {
		endpoint = c.Endpoint[7:]
	} else {
		endpoint = c.Endpoint
	}

	// SLS public request headers
	var hostStr string
	if len(project) == 0 {
		hostStr = endpoint
	} else {
		hostStr = project + "." + endpoint
	}
	headers[HTTPHeaderHost] = hostStr
	headers[HTTPHeaderAPIVersion] = version

	if len(c.UserAgent) > 0 {
		headers[HTTPHeaderUserAgent] = c.UserAgent
	} else {
		headers[HTTPHeaderUserAgent] = DefaultLogUserAgent
	}

	c.accessKeyLock.RLock()
	stsToken := c.SecurityToken
	accessKeyID := c.AccessKeyID
	accessKeySecret := c.AccessKeySecret
	region := c.Region
	authVersion := c.AuthVersion
	c.accessKeyLock.RUnlock()

	if c.credentialsProvider != nil {
		res, err := c.credentialsProvider.GetCredentials()
		if err != nil {
			return nil, fmt.Errorf("fail to fetch credentials: %w", err)
		}
		accessKeyID = res.AccessKeyID
		accessKeySecret = res.AccessKeySecret
		stsToken = res.SecurityToken
	}

	// Access with token
	if stsToken != "" {
		headers[HTTPHeaderAcsSecurityToken] = stsToken
	}

	if body != nil {
		if _, ok := headers[HTTPHeaderContentType]; !ok {
			return nil, fmt.Errorf("Can't find 'Content-Type' header")
		}
	}
	for k, v := range c.InnerHeaders {
		headers[k] = v
	}
	var signer Signer
	if authVersion == AuthV4 {
		headers[HTTPHeaderLogDate] = dateTimeISO8601()
		signer = NewSignerV4(accessKeyID, accessKeySecret, region)
	} else if authVersion == AuthV0 {
		signer = NewSignerV0()
	} else {
		headers[HTTPHeaderDate] = nowRFC1123()
		signer = NewSignerV1(accessKeyID, accessKeySecret)
	}
	if err := signer.Sign(method, uri, headers, body); err != nil {
		return nil, err
	}

	addHeadersAfterSign(c.CommonHeaders, headers)
	// Initialize http request
	reader := bytes.NewReader(body)
	var urlStr string
	// using http as default
	if !GlobalForceUsingHTTP && usingHTTPS {
		urlStr = "https://"
	} else {
		urlStr = "http://"
	}
	urlStr += hostStr + uri
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	if IsDebugLevelMatched(5) {
		dump, e := httputil.DumpRequest(req, true)
		if e != nil {
			level.Info(Logger).Log("msg", e)
		}
		level.Info(Logger).Log("msg", "HTTP Request:\n%v", string(dump))
	}

	// Get ready to do request
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = defaultHttpClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse the sls error from body.
	if resp.StatusCode != http.StatusOK {
		err := &Error{}
		err.HTTPCode = (int32)(resp.StatusCode)
		defer resp.Body.Close()
		buf, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(buf, err)
		err.RequestID = resp.Header.Get("x-log-requestid")
		return nil, err
	}
	if IsDebugLevelMatched(5) {
		dump, e := httputil.DumpResponse(resp, true)
		if e != nil {
			level.Info(Logger).Log("msg", e)
		}
		level.Info(Logger).Log("msg", "HTTP Response:\n%v", string(dump))
	}

	return resp, nil
}
