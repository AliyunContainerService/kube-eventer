package sls

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-kit/kit/log/level"
	"golang.org/x/net/context"
)

// timeout configs
var (
	defaultRequestTimeout  = 60 * time.Second
	defaultRetryTimeout    = 90 * time.Second
	defaultHttpClient      = newDefaultHTTPClient(defaultRequestTimeout)
	defaultHTTPIdleTimeout = time.Second * 55
)

func newDefaultTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.IdleConnTimeout = defaultHTTPIdleTimeout
	return t
}

// returns a new http client instance with default config
func newDefaultHTTPClient(requestTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: newDefaultTransport(),
		Timeout:   requestTimeout,
	}
}

func retryReadErrorCheck(ctx context.Context, err error) (bool, error) {
	if err == nil {
		return false, nil
	}
	switch e := err.(type) {
	case *url.Error:
		return true, e
	case *Error:
		if RetryOnServerErrorEnabled {
			if e.HTTPCode >= 500 && e.HTTPCode <= 599 {
				return true, e
			}
		}
	case *BadResponseError:
		if RetryOnServerErrorEnabled {
			if e.HTTPCode >= 500 && e.HTTPCode <= 599 {
				return true, e
			}
		}
	default:
		return false, e
	}

	return false, err
}

func retryWriteErrorCheck(ctx context.Context, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	switch e := err.(type) {
	case *Error:
		if RetryOnServerErrorEnabled {
			if e.HTTPCode == 500 || e.HTTPCode == 502 || e.HTTPCode == 503 {
				return true, e
			}
		}
	case *BadResponseError:
		if RetryOnServerErrorEnabled {
			if e.HTTPCode == 500 || e.HTTPCode == 502 || e.HTTPCode == 503 {
				return true, e
			}
		}
	default:
		return false, e
	}

	return false, err
}

// request sends a request to SLS.
// mock param only for test, default is []
func request(project *LogProject, method, uri string, headers map[string]string,
	body []byte, mock ...interface{}) (*http.Response, error) {

	var r *http.Response
	var slsErr error
	var err error
	var mockErr *mockErrorRetry

	project.init()
	ctx, cancel := context.WithTimeout(context.Background(), project.retryTimeout)
	defer cancel()

	//fmt.Println("request ", project, method, uri, headers, body)
	// all GET method is read function
	if method == http.MethodGet {
		err = RetryWithCondition(ctx, backoff.NewExponentialBackOff(), func() (bool, error) {
			if len(mock) == 0 {
				//fmt.Println("real request", project, method, uri, headers, body)
				r, slsErr = realRequest(ctx, project, method, uri, headers, body)
				//fmt.Println("real request done")
			} else {
				r, mockErr = nil, mock[0].(*mockErrorRetry)
				mockErr.RetryCnt--
				if mockErr.RetryCnt <= 0 {
					r = &http.Response{}
					slsErr = nil
					return false, nil
				}
				slsErr = &mockErr.Err
			}
			return retryReadErrorCheck(ctx, slsErr)
		})
	} else {
		err = RetryWithCondition(ctx, backoff.NewExponentialBackOff(), func() (bool, error) {
			if len(mock) == 0 {
				r, slsErr = realRequest(ctx, project, method, uri, headers, body)
			} else {
				r, mockErr = nil, mock[0].(*mockErrorRetry)
				mockErr.RetryCnt--
				if mockErr.RetryCnt <= 0 {
					r = &http.Response{}
					slsErr = nil
					return false, nil
				}
				slsErr = &mockErr.Err
			}
			return retryWriteErrorCheck(ctx, slsErr)
		})
	}

	if err != nil {
		return r, err
	}
	return r, slsErr
}

// request sends a request to alibaba cloud Log Service.
// @note if error is nil, you must call http.Response.Body.Close() to finalize reader
func realRequest(ctx context.Context, project *LogProject, method, uri string, headers map[string]string,
	body []byte) (*http.Response, error) {

	// The caller should provide 'x-log-bodyrawsize' header
	if _, ok := headers[HTTPHeaderBodyRawSize]; !ok {
		return nil, NewClientError(fmt.Errorf("Can't find 'x-log-bodyrawsize' header"))
	}

	// SLS public request headers
	baseURL := project.getBaseURL()
	headers[HTTPHeaderHost] = baseURL
	headers[HTTPHeaderAPIVersion] = version
	if len(project.UserAgent) > 0 {
		headers[HTTPHeaderUserAgent] = project.UserAgent
	} else {
		headers[HTTPHeaderUserAgent] = DefaultLogUserAgent
	}

	stsToken := project.SecurityToken
	accessKeyID := project.AccessKeyID
	accessKeySecret := project.AccessKeySecret

	if project.credentialProvider != nil {
		c, err := project.credentialProvider.GetCredentials()
		if err != nil {
			return nil, NewClientError(fmt.Errorf("fail to get credentials: %w", err))
		}
		stsToken = c.SecurityToken
		accessKeyID = c.AccessKeyID
		accessKeySecret = c.AccessKeySecret
	}

	// Access with token
	if stsToken != "" {
		headers[HTTPHeaderAcsSecurityToken] = stsToken
	}

	if body != nil {
		if _, ok := headers[HTTPHeaderContentType]; !ok {
			return nil, NewClientError(fmt.Errorf("Can't find 'Content-Type' header"))
		}
	}

	for k, v := range project.InnerHeaders {
		headers[k] = v
	}
	var signer Signer
	if project.AuthVersion == AuthV4 {
		headers[HTTPHeaderLogDate] = dateTimeISO8601()
		signer = NewSignerV4(accessKeyID, accessKeySecret, project.Region)
	} else if project.AuthVersion == AuthV0 {
		signer = NewSignerV0()
	} else {
		headers[HTTPHeaderDate] = nowRFC1123()
		signer = NewSignerV1(accessKeyID, accessKeySecret)
	}
	if err := signer.Sign(method, uri, headers, body); err != nil {
		return nil, err
	}

	addHeadersAfterSign(project.CommonHeaders, headers)

	// Initialize http request
	reader := bytes.NewReader(body)

	// Handle the endpoint
	urlStr := fmt.Sprintf("%s%s", baseURL, uri)
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, NewClientError(err)
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
	resp, err := project.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse the sls error from body.
	if resp.StatusCode != http.StatusOK {
		err := &Error{}
		err.HTTPCode = (int32)(resp.StatusCode)
		defer resp.Body.Close()
		buf, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			return nil, NewBadResponseError(ioErr.Error(), resp.Header, resp.StatusCode)
		}
		if jErr := json.Unmarshal(buf, err); jErr != nil {
			return nil, NewBadResponseError(string(buf), resp.Header, resp.StatusCode)
		}
		err.RequestID = resp.Header.Get(RequestIDHeader)
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
