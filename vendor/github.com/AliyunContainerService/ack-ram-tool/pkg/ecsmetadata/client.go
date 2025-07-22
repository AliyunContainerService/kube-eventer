package ecsmetadata

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// https://help.aliyun.com/zh/ecs/user-guide/view-instance-metadata

const (
	DefaultEndpoint        = "http://100.100.100.200"
	EnvEndpoint            = "ALIBABA_CLOUD_IMDS_ENDPOINT"
	EnvIMDSV2Disabled      = "ALIBABA_CLOUD_IMDSV2_DISABLED"
	EnvIMDSRoleName        = "ALIBABA_CLOUD_ECS_METADATA"
	defaultTokenTTLSeconds = 18000 // 5hours
	minTokenTTLSeconds     = 1
	maxTokenTTLSeconds     = 21600
	defaultClientTimeout   = time.Second * 30
)

var DefaultClient, _ = NewClient(ClientOptions{})
var UserAgent = ""

type Client struct {
	httpClient *http.Client

	endpoint        string
	roleName        string
	disableIMDSV2   bool
	tokenTTLSeconds int

	metadataToken    string
	metadataTokenExp time.Time

	nowFunc func() time.Time

	disableRetry bool
	retryOptions RetryOptions

	userAgent string
}

type TransportWrapper func(rt http.RoundTripper) http.RoundTripper

type ClientOptions struct {
	// default: DefaultEndpoint
	Endpoint string
	// ram role of ecs instance
	RoleName      string
	DisableIMDSV2 bool
	// default: 18000, 5 hours
	TokenTTLSeconds int

	TransportWrappers []TransportWrapper
	transport         http.RoundTripper
	// default: 30 seconds
	Timeout time.Duration
	// default: time.Now
	NowFunc func() time.Time

	DisableRetry bool
	// default: DefaultRetryOptions()
	RetryOptions *RetryOptions

	UserAgent string
}

func init() {
	name := path.Base(os.Args[0])
	UserAgent = fmt.Sprintf("%s %s/%s %s", name, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func NewClient(opts ClientOptions) (*Client, error) {
	if err := opts.prepare(); err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Transport: opts.transport,
		Timeout:   opts.Timeout,
	}
	return &Client{
		httpClient:      httpClient,
		endpoint:        opts.Endpoint,
		roleName:        opts.RoleName,
		disableIMDSV2:   opts.DisableIMDSV2,
		tokenTTLSeconds: opts.TokenTTLSeconds,
		nowFunc:         opts.NowFunc,
		retryOptions:    *opts.RetryOptions,
		disableRetry:    opts.DisableRetry,
		userAgent:       opts.UserAgent,
	}, nil
}

func (c *Client) GetMetaData(ctx context.Context, method, path string) ([]byte, error) {
	return c.getMetaDataWithRetry(ctx, method, path)
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	if c.disableIMDSV2 {
		return "", nil
	}

	now := c.getNow()
	if c.metadataToken != "" && !c.tokenExpired(now) {
		return c.metadataToken, nil
	}

	h := http.Header{}
	h.Set("X-aliyun-ecs-metadata-token-ttl-seconds", fmt.Sprintf("%d", c.tokenTTLSeconds))
	body, err := c.send(ctx, http.MethodPut, "/latest/api/token", h)
	if err != nil {
		return "", fmt.Errorf("get token failed: %w", err)
	}

	c.metadataToken = strings.TrimSpace(string(body))
	c.metadataTokenExp = now.
		Add(time.Duration(float64(c.tokenTTLSeconds)*0.8) * time.Second).
		Add(-time.Minute)

	return c.metadataToken, nil
}

func (c *Client) getMetaDataWithRetry(ctx context.Context, method, path string) ([]byte, error) {
	if c.disableRetry {
		return c.getMetaData(ctx, method, path)
	}

	var data []byte
	var err error
	lastErr := retryWithOptions(ctx, func(ctx context.Context) error {
		data, err = c.getMetaData(ctx, method, path)
		if err != nil {
			if !isRetryable(err) {
				return newNoRetryError(err)
			}
		}
		return err
	}, c.retryOptions)

	return data, lastErr
}

func (c *Client) getMetaData(ctx context.Context, method, path string) ([]byte, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		var httpErr *HTTPError
		if errors.As(err, &httpErr) &&
			(httpErr.StatusCode == http.StatusNotFound ||
				httpErr.StatusCode == http.StatusForbidden) {
			// ignore 404 and 403 error
		} else {
			return nil, err
		}
	}

	h := http.Header{}
	if token != "" {
		h.Set("X-aliyun-ecs-metadata-token", token)
	}
	return c.send(ctx, method, path, h)
}

func (c *Client) getTidyStringData(ctx context.Context, path string) (string, error) {
	data, err := c.getRawStringData(ctx, path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(data), nil
}

func (c *Client) getRawStringData(ctx context.Context, path string) (string, error) {
	data, err := c.GetMetaData(ctx, http.MethodGet, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) getRawData(ctx context.Context, path string) ([]byte, error) {
	return c.GetMetaData(ctx, http.MethodGet, path)
}

func (c *Client) send(ctx context.Context, method, path string, header http.Header) ([]byte, error) {
	url := c.getURL(path)
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	for k, v := range header {
		req.Header.Set(k, v[0])
	}
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		err2 := fmt.Errorf("do request failed: %w", err)
		return nil, newHTTPError(err2, url, resp, nil)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err2 := fmt.Errorf("read body failed: %w", err)
		return nil, newHTTPError(err2, url, resp, nil)
	}
	if resp.StatusCode != http.StatusOK {
		err2 := fmt.Errorf("status code of respose is not 200: %s", resp.Status)
		return nil, newHTTPError(err2, url, resp, body)
	}
	return body, nil
}

func (c *Client) tokenExpired(now time.Time) bool {
	if c.metadataTokenExp.IsZero() {
		return true
	}
	return c.metadataTokenExp.Before(now)
}

func (c *Client) getNow() time.Time {
	if c.nowFunc != nil {
		return c.nowFunc()
	}
	return time.Now()
}

func (o *ClientOptions) prepare() error {
	if o.Timeout <= 0 {
		o.Timeout = defaultClientTimeout
	}
	if o.transport == nil {
		ts := http.DefaultTransport.(*http.Transport).Clone()
		o.transport = ts
	}
	if len(o.TransportWrappers) > 0 {
		for _, tw := range o.TransportWrappers {
			o.transport = tw(o.transport)
		}
	}
	if o.Endpoint == "" {
		if v := os.Getenv(EnvEndpoint); v != "" {
			o.Endpoint = v
		} else {
			o.Endpoint = DefaultEndpoint
		}
	} else {
		o.Endpoint = strings.TrimRight(o.Endpoint, "/")
	}
	if !o.DisableIMDSV2 {
		if v := os.Getenv(EnvIMDSV2Disabled); v != "" {
			if b, err := strconv.ParseBool(v); err == nil && b {
				o.DisableIMDSV2 = true
			}
		}
	}
	if o.TokenTTLSeconds == 0 {
		o.TokenTTLSeconds = defaultTokenTTLSeconds
	}
	if o.TokenTTLSeconds < minTokenTTLSeconds || o.TokenTTLSeconds > maxTokenTTLSeconds {
		return fmt.Errorf("invalid TokenTTLSeconds: %d", o.TokenTTLSeconds)
	}
	if o.RoleName == "" {
		if v := os.Getenv(EnvIMDSRoleName); v != "" {
			o.RoleName = v
		}
	}

	if !o.DisableRetry {
		if o.RetryOptions == nil {
			o.RetryOptions = DefaultRetryOptions()
		}
	}
	if o.UserAgent == "" {
		o.UserAgent = UserAgent
	}

	return nil
}

func (c *Client) SetEndpoint(endpoint string) *Client {
	c.endpoint = strings.TrimRight(endpoint, "/")
	return c
}

func (c *Client) getURL(path string) string {
	return fmt.Sprintf("%s%s", c.endpoint, path)
}
