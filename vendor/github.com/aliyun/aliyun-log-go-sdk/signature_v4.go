package sls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	emptyStringSha256                = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	signerV4ProductName              = "sls"
	ISO8601                          = "20060102T150405Z"
	authorizationV4SigningSaltFigure = "aliyun_v4_request"
	authorizationAlgorithmV4         = "SLS4-HMAC-SHA256"
	authorizationV4SecretKeyPrefix   = "aliyun_v4"
)

var (
	errSignerV4MissingRegion = errors.New("sign version v4 require a valid region")
)

// SignerV4 sign version v4, a non-empty region is required
type SignerV4 struct {
	accessKeyID     string
	accessKeySecret string
	region          string
}

func NewSignerV4(accessKeyID, accessKeySecret, region string) *SignerV4 {
	return &SignerV4{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		region:          region,
	}
}

func (s *SignerV4) isSignedHeader(key string) bool {
	return strings.HasPrefix(key, "x-log-") ||
		strings.HasPrefix(key, "x-acs-") ||
		strings.EqualFold(key, HTTPHeaderHost) ||
		strings.EqualFold(key, HTTPHeaderContentType)
}

func (s *SignerV4) Sign(method, uri string, headers map[string]string, body []byte) error {
	if s.region == "" {
		return errSignerV4MissingRegion
	}

	uri, urlParams, err := s.parseUri(uri)
	if err != nil {
		return err
	}

	dateTime, ok := headers[HTTPHeaderLogDate]
	if !ok {
		return fmt.Errorf("can't find '%s' header", HTTPHeaderLogDate)
	}
	date := dateTime[:8]
	// Host should not contain schema here.
	if host, ok := headers[HTTPHeaderHost]; ok {
		if strings.HasPrefix(host, "http://") {
			headers[HTTPHeaderHost] = host[len("http://"):]
		} else if strings.HasPrefix(host, "https://") {
			headers[HTTPHeaderHost] = host[len("https://"):]
		}
	}

	contentLength := len(body)
	var sha256Payload string
	if contentLength != 0 {
		sha256Payload = fmt.Sprintf("%x", sha256.Sum256(body))
	} else {
		sha256Payload = emptyStringSha256
	}
	headers[HTTPHeaderLogContentSha256] = sha256Payload
	headers[HTTPHeaderContentLength] = strconv.Itoa(contentLength)

	// Canonical headers
	signedHeadersStr, canonicalHeaderStr := s.buildCanonicalHeaders(headers)

	// CanonicalRequest
	canonReq := s.buildCanonicalRequest(method, uri, sha256Payload, canonicalHeaderStr, signedHeadersStr, urlParams)
	scope := s.buildScope(date, s.region)

	// SignKey + signMessage => signature
	strToSign := s.buildSignMessage(canonReq, dateTime, scope)
	key, err := s.buildSigningKey(s.accessKeySecret, s.region, date)
	if err != nil {
		return err
	}
	hash, err := s.hmacSha256([]byte(strToSign), key)
	if err != nil {
		return err
	}
	signature := hex.EncodeToString(hash)
	headers[HTTPHeaderAuthorization] = s.buildAuthorization(s.accessKeyID, signature, scope)
	return nil
}

func (s *SignerV4) buildCanonicalHeaders(headers map[string]string) (string, string) {
	var headerKeys []string
	signed := make(map[string]string)
	for k, v := range headers {
		key := strings.ToLower(k)
		if s.isSignedHeader(key) {
			signed[key] = v
			headerKeys = append(headerKeys, key)
		}
	}
	sort.Strings(headerKeys)
	var canonicalHeaders strings.Builder
	var signedHeaders strings.Builder
	n := len(headerKeys)
	for i := 0; i < n; i++ {
		canonicalHeaders.WriteString(headerKeys[i])
		canonicalHeaders.WriteRune(':')
		canonicalHeaders.WriteString(signed[headerKeys[i]])
		canonicalHeaders.WriteRune('\n')
		if i > 0 {
			signedHeaders.WriteRune(';')
		}
		signedHeaders.WriteString(headerKeys[i])
	}
	return signedHeaders.String(), canonicalHeaders.String()
}

func (s *SignerV4) parseUri(uriWithQuery string) (string, map[string]string, error) {
	u, err := url.Parse(uriWithQuery)
	if err != nil {
		return "", nil, err
	}
	urlParams := make(map[string]string)
	for k, vals := range u.Query() {
		if len(vals) == 0 {
			urlParams[k] = ""
		} else {
			urlParams[k] = vals[0] // param val should at most one value
		}
	}
	return u.Path, urlParams, nil
}

func dateTimeISO8601() string {
	return time.Now().In(gmtLoc).Format(ISO8601)
}

func (s *SignerV4) buildCanonicalRequest(method, uri, sha256Payload, canonicalHeaders, signedHeaders string, urlParams map[string]string) string {
	builder := strings.Builder{}
	builder.WriteString(method)
	builder.WriteRune('\n')
	builder.WriteString(uri)
	builder.WriteRune('\n')

	// Url params
	canonParams := make(map[string]string)
	var queryKeys []string
	for k, v := range urlParams {
		canonParams[k] = s.percentEncode(v)
		queryKeys = append(queryKeys, k)
	}
	sort.Strings(queryKeys)
	n := len(queryKeys)
	for i := 0; i < n; i++ {
		if i > 0 {
			builder.WriteRune('&')
		}
		builder.WriteString(queryKeys[i])
		v := canonParams[queryKeys[i]]
		if len(v) != 0 {
			builder.WriteRune('=')
			builder.WriteString(v)
		}
	}
	builder.WriteRune('\n')
	builder.WriteString(canonicalHeaders)
	builder.WriteRune('\n')
	builder.WriteString(signedHeaders)
	builder.WriteRune('\n')
	builder.WriteString(sha256Payload)
	return builder.String()
}

func (s *SignerV4) percentEncode(uri string) string {
	u := url.QueryEscape(uri)
	u = strings.ReplaceAll(u, "+", "%20")
	return u
}

func (s *SignerV4) buildScope(date, region string) string {
	var builder strings.Builder
	builder.WriteString(date)
	builder.WriteRune('/')
	builder.WriteString(region)
	builder.WriteRune('/')
	builder.WriteString(signerV4ProductName)
	builder.WriteRune('/')
	builder.WriteString(authorizationV4SigningSaltFigure)
	return builder.String()
}

func (s *SignerV4) buildSignMessage(canonReq, dateTime, scope string) string {
	var builder strings.Builder
	builder.WriteString(authorizationAlgorithmV4)
	builder.WriteRune('\n')
	builder.WriteString(dateTime)
	builder.WriteRune('\n')
	builder.WriteString(scope)
	builder.WriteRune('\n')
	builder.WriteString(fmt.Sprintf("%x", sha256.Sum256([]byte(canonReq))))
	return builder.String()
}

func (s *SignerV4) hmacSha256(message, key []byte) ([]byte, error) {
	hmacHasher := hmac.New(sha256.New, key)
	_, err := hmacHasher.Write(message)
	if err != nil {
		return nil, err
	}
	return hmacHasher.Sum(nil), nil
}

func (s *SignerV4) buildSigningKey(accessKeySecret, region, date string) ([]byte, error) {
	signDate, err := s.hmacSha256([]byte(date), []byte(authorizationV4SecretKeyPrefix+accessKeySecret))
	if err != nil {
		return nil, err
	}
	signRegion, err := s.hmacSha256([]byte(region), signDate)
	if err != nil {
		return nil, err
	}
	signService, err := s.hmacSha256([]byte(signerV4ProductName), signRegion)
	if err != nil {
		return nil, err
	}
	signAll, err := s.hmacSha256([]byte(authorizationV4SigningSaltFigure), signService)
	if err != nil {
		return nil, err
	}
	return signAll, nil
}

func (s *SignerV4) buildAuthorization(accessKeyID, signature, scope string) string {
	return fmt.Sprintf("SLS4-HMAC-SHA256 Credential=%s/%s,Signature=%s", accessKeyID, scope, signature)
}
