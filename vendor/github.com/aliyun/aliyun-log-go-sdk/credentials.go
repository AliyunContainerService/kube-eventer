package sls

import (
	"time"
)

type Credentials struct {
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
}

const DEFAULT_EXPIRED_FACTOR = 0.8

// Expirable credentials with an expiration.
type TempCredentials struct {
	Credentials
	expiredFactor      float64
	expirationInMills  int64 // The time when the credentials expires, unix timestamp in millis
	lastUpdatedInMills int64
}

func NewTempCredentials(accessKeyId, accessKeySecret, securityToken string,
	expirationInMills, lastUpdatedInMills int64) *TempCredentials {

	return &TempCredentials{
		Credentials: Credentials{
			AccessKeyID:     accessKeyId,
			AccessKeySecret: accessKeySecret,
			SecurityToken:   securityToken,
		},
		expirationInMills:  expirationInMills,
		lastUpdatedInMills: lastUpdatedInMills,
		expiredFactor:      DEFAULT_EXPIRED_FACTOR,
	}
}

// @param factor must > 0.0 and <= 1.0, the less the factor is,
// the more frequently the credentials will be updated.
//
// If factor is set to 0, the credentials will be fetched every time
// [GetCredentials] is called.
//
// If factor is set to 1, the credentials will be fetched only when expired .
func (t *TempCredentials) WithExpiredFactor(factor float64) *TempCredentials {
	if factor > 0.0 && factor <= 1.0 {
		t.expiredFactor = factor
	}
	return t
}

// Returns true if credentials has expired already or will expire soon.
func (t *TempCredentials) ShouldRefresh() bool {
	nowInMills := time.Now().UnixNano() / 1e6
	if nowInMills >= t.expirationInMills {
		return true
	}
	duration := (float64)(t.expirationInMills-t.lastUpdatedInMills) * t.expiredFactor
	if duration < 0.0 { // check here
		duration = 0
	}
	return (nowInMills - t.lastUpdatedInMills) >= int64(duration)
}

// Returns true if credentials has expired already.
func (t *TempCredentials) HasExpired() bool {
	nowInMills := time.Now().UnixNano() / 1e6
	return nowInMills >= t.expirationInMills
}
