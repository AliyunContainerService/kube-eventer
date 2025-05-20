package sls

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"go.uber.org/atomic"

	"github.com/go-kit/kit/log/level"
)

const CRED_TIME_FORMAT = time.RFC3339

type CredentialsProvider interface {
	/**
	 * GetCredentials is called everytime credentials are needed, the CredentialsProvider
	 * should cache credentials to avoid fetch credentials too frequently.
	 *
	 * @note GetCredentials must be thread-safe to avoid data race.
	 */
	GetCredentials() (Credentials, error)
}

/**
 * A static credetials provider that always returns the same long-lived credentials.
 * For back compatible.
 */
type StaticCredentialsProvider struct {
	Cred Credentials
}

// Create a static credential provider with AccessKeyID/AccessKeySecret/SecurityToken.
//
// Param accessKeyID and accessKeySecret must not be an empty string.
func NewStaticCredentialsProvider(accessKeyID, accessKeySecret, securityToken string) *StaticCredentialsProvider {
	return &StaticCredentialsProvider{
		Cred: Credentials{
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
			SecurityToken:   securityToken,
		},
	}
}

func (p *StaticCredentialsProvider) GetCredentials() (Credentials, error) {
	return p.Cred, nil
}

type CredentialsRequestBuilder = func() (*http.Request, error)
type CredentialsRespParser = func(*http.Response) (*TempCredentials, error)
type CredentialsFetcher = func() (*TempCredentials, error)

// Combine RequestBuilder and RespParser, return a CredentialsFetcher
func NewCredentialsFetcher(builder CredentialsRequestBuilder, parser CredentialsRespParser, customClient *http.Client) CredentialsFetcher {
	return func() (*TempCredentials, error) {
		req, err := builder()
		if err != nil {
			return nil, fmt.Errorf("fail to build http request: %w", err)
		}

		var client *http.Client
		if customClient != nil {
			client = customClient
		} else {
			client = &http.Client{}
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fail to do http request: %w", err)
		}
		defer resp.Body.Close()
		cred, err := parser(resp)
		if err != nil {
			return nil, fmt.Errorf("fail to parse http response: %w", err)
		}
		return cred, nil
	}
}

// Wraps a CredentialsFetcher with retry.
//
// @param retryTimes If <= 0, no retry will be performed.
func fetcherWithRetry(fetcher CredentialsFetcher, retryTimes int) CredentialsFetcher {
	return func() (*TempCredentials, error) {
		var errs []error
		for i := 0; i <= retryTimes; i++ {
			cred, err := fetcher()
			if err == nil {

				return cred, nil
			}
			errs = append(errs, err)
		}
		return nil, fmt.Errorf("exceed max retry times, errors: %w",
			joinErrors(errs...))
	}
}

// Replace this with errors.Join when go version >= 1.20
func joinErrors(errs ...error) error {
	if errs == nil {
		return nil
	}
	errStrs := make([]string, 0, len(errs))
	for _, e := range errs {
		errStrs = append(errStrs, e.Error())
	}
	return fmt.Errorf("[%s]", strings.Join(errStrs, ", "))
}

const UPDATE_FUNC_RETRY_TIMES = 3
const UPDATE_FUNC_FETCH_ADVANCED_DURATION = time.Second * 60 * 10

// Adapter for porting UpdateTokenFunc to a CredentialsProvider.
type UpdateFuncProviderAdapter struct {
	cred atomic.Value // type *Credentials

	fetcher CredentialsFetcher

	expirationInMills atomic.Int64
	advanceDuration   time.Duration // fetch before credentials expires in advance
}

// Returns a new CredentialsProvider.
func NewUpdateFuncProviderAdapter(updateFunc UpdateTokenFunction) *UpdateFuncProviderAdapter {
	retryTimes := UPDATE_FUNC_RETRY_TIMES
	fetcher := fetcherWithRetry(updateFuncFetcher(updateFunc), retryTimes)

	return &UpdateFuncProviderAdapter{
		advanceDuration: UPDATE_FUNC_FETCH_ADVANCED_DURATION,
		fetcher:         fetcher,
	}
}

func updateFuncFetcher(updateFunc UpdateTokenFunction) CredentialsFetcher {
	return func() (*TempCredentials, error) {
		id, secret, token, expireTime, err := updateFunc()
		if err != nil {
			return nil, fmt.Errorf("updateTokenFunc fetch credentials failed: %w", err)
		}

		if !checkSTSTokenValid(id, secret, token, expireTime) {
			return nil, fmt.Errorf("updateTokenFunc result not valid, expirationTime:%s",
				expireTime.Format(time.RFC3339))
		}
		return NewTempCredentials(id, secret, token, expireTime.UnixNano()/1e6, -1), nil
	}

}

// If credentials expires or will be exipred soon, fetch a new credentials and return it.
//
// Otherwise returns the credentials fetched last time.
//
// Retry at most maxRetryTimes if failed to fetch.
func (adp *UpdateFuncProviderAdapter) GetCredentials() (Credentials, error) {
	if !adp.shouldRefresh() {
		res := adp.cred.Load().(*Credentials)
		return *res, nil
	}
	level.Debug(Logger).Log("reason", "updateTokenFunc start to fetch new credentials")

	res, err := adp.fetcher() // res.lastUpdatedTime is not valid, do not use it

	if err != nil {
		return Credentials{}, fmt.Errorf("updateTokenFunc fail to fetch credentials, err:%w", err)
	}

	adp.cred.Store(&res.Credentials)
	adp.expirationInMills.Store(res.expirationInMills)
	level.Debug(Logger).Log("reason", "updateTokenFunc fetch new credentials succeed",
		"expirationTime", time.Unix(res.expirationInMills/1e3, res.expirationInMills%1e3*1e6).Format(CRED_TIME_FORMAT),
	)
	return res.Credentials, nil
}

// Returns true if no credentials ever fetched or credentials expired,
// or credentials will be expired soon
func (adp *UpdateFuncProviderAdapter) shouldRefresh() bool {
	v := adp.cred.Load()
	if v == nil {
		return true
	}
	now := time.Now()
	millis := adp.expirationInMills.Load()
	return time.Unix(millis/1e3, millis%1e3*1e6).Sub(now) <= adp.advanceDuration
}

func checkSTSTokenValid(accessKeyID, accessKeySecret, securityToken string, expirationTime time.Time) bool {
	return accessKeyID != "" && accessKeySecret != "" && expirationTime.UnixNano() > 0
}

const ECS_RAM_ROLE_URL_PREFIX = "http://100.100.100.200/latest/meta-data/ram/security-credentials/"
const ECS_RAM_ROLE_RETRY_TIMES = 3

func NewEcsRamRoleFetcher(urlPrefix, ramRole string, customClient *http.Client) CredentialsFetcher {
	return NewCredentialsFetcher(newEcsRamRoleReqBuilder(urlPrefix, ramRole),
		ecsRamRoleParser, customClient)
}

// Build http GET request with url(urlPrefix + ramRole)
func newEcsRamRoleReqBuilder(urlPrefix, ramRole string) func() (*http.Request, error) {
	return func() (*http.Request, error) {
		url := urlPrefix + ramRole
		return http.NewRequest(http.MethodGet, url, nil)
	}
}

// Parse ECS Ram Role http response, convert it to TempCredentials
func ecsRamRoleParser(resp *http.Response) (*TempCredentials, error) {
	// 1. read body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read http resp body: %w", err)
	}
	fetchResp := EcsRamRoleHttpResp{}
	// 2. unmarshal
	err = json.Unmarshal(data, &fetchResp)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal json: %w, body: %s", err, string(data))
	}
	// 3. check json param
	if !fetchResp.isValid() {
		return nil, fmt.Errorf("invalid fetch result, body: %s", string(data))
	}
	return NewTempCredentials(
		fetchResp.AccessKeyID,
		fetchResp.AccessKeySecret,
		fetchResp.SecurityToken, fetchResp.Expiration, fetchResp.LastUpdated), nil
}

// Response struct for http response of ecs ram role fetch request
type EcsRamRoleHttpResp struct {
	Code            string `json:"Code"`
	AccessKeyID     string `json:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret"`
	SecurityToken   string `json:"SecurityToken"`
	Expiration      int64  `json:"Expiration"`
	LastUpdated     int64  `json:"LastUpdated"`
}

func (r *EcsRamRoleHttpResp) isValid() bool {
	return strings.ToLower(r.Code) == "success" && r.AccessKeyID != "" &&
		r.AccessKeySecret != "" && r.Expiration > 0 && r.LastUpdated > 0
}
