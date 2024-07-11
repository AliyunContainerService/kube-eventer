package sls

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	HTTPHeaderAuthorization    = "Authorization"
	HTTPHeaderContentMD5       = "Content-MD5"
	HTTPHeaderContentType      = "Content-Type"
	HTTPHeaderContentLength    = "Content-Length"
	HTTPHeaderDate             = "Date"
	HTTPHeaderHost             = "Host"
	HTTPHeaderUserAgent        = "User-Agent"
	HTTPHeaderAcsSecurityToken = "x-acs-security-token"
	HTTPHeaderAPIVersion       = "x-log-apiversion"
	HTTPHeaderLogDate          = "x-log-date"
	HTTPHeaderLogContentSha256 = "x-log-content-sha256"
	HTTPHeaderSignatureMethod  = "x-log-signaturemethod"
	HTTPHeaderBodyRawSize      = "x-log-bodyrawsize"
)

type Signer interface {
	// Sign modifies @param headers only, adds signature and other http headers
	// that log services authorization requires.
	Sign(method, uriWithQuery string, headers map[string]string, body []byte) error
}

// GMT location
var gmtLoc = time.FixedZone("GMT", 0)

// NowRFC1123 returns now time in RFC1123 format with GMT timezone,
// eg, "Mon, 02 Jan 2006 15:04:05 GMT".
func nowRFC1123() string {
	return time.Now().In(gmtLoc).Format(time.RFC1123)
}
func NewSignerV0() *SignerV0 {
	return &SignerV0{}
}

type SignerV0 struct{}

func (s *SignerV0) Sign(method, uriWithQuery string, headers map[string]string, body []byte) error {
	// do nothing
	return nil
}

// SignerV1 version v1
type SignerV1 struct {
	accessKeyID     string
	accessKeySecret string
}

func NewSignerV1(accessKeyID, accessKeySecret string) *SignerV1 {
	return &SignerV1{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
	}
}

func (s *SignerV1) Sign(method, uri string, headers map[string]string, body []byte) error {
	var contentMD5, contentType, date, canoHeaders, canoResource string
	if body != nil {
		contentMD5 = fmt.Sprintf("%X", md5.Sum(body))
		headers[HTTPHeaderContentMD5] = contentMD5
	}

	if val, ok := headers[HTTPHeaderContentType]; ok {
		contentType = val
	}

	date, ok := headers[HTTPHeaderDate]
	if !ok {
		return fmt.Errorf("Can't find 'Date' header")
	}
	headers[HTTPHeaderSignatureMethod] = signatureMethod
	var slsHeaderKeys sort.StringSlice

	// Calc CanonicalizedSLSHeaders
	slsHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		l := strings.TrimSpace(strings.ToLower(k))
		if strings.HasPrefix(l, "x-log-") || strings.HasPrefix(l, "x-acs-") {
			slsHeaders[l] = strings.TrimSpace(v)
			slsHeaderKeys = append(slsHeaderKeys, l)
		}
	}

	sort.Sort(slsHeaderKeys)
	for i, k := range slsHeaderKeys {
		canoHeaders += k + ":" + slsHeaders[k]
		if i+1 < len(slsHeaderKeys) {
			canoHeaders += "\n"
		}
	}

	// Calc CanonicalizedResource
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	canoResource += u.EscapedPath()
	if u.RawQuery != "" {
		var keys sort.StringSlice

		vals := u.Query()
		for k := range vals {
			keys = append(keys, k)
		}

		sort.Sort(keys)
		canoResource += "?"
		for i, k := range keys {
			if i > 0 {
				canoResource += "&"
			}

			for _, v := range vals[k] {
				canoResource += k + "=" + v
			}
		}
	}

	signStr := method + "\n" +
		contentMD5 + "\n" +
		contentType + "\n" +
		date + "\n" +
		canoHeaders + "\n" +
		canoResource

	// Signature = base64(hmac-sha1(UTF8-Encoding-Of(SignString)，AccessKeySecret))
	mac := hmac.New(sha1.New, []byte(s.accessKeySecret))
	_, err = mac.Write([]byte(signStr))
	if err != nil {
		return err
	}
	digest := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	auth := fmt.Sprintf("SLS %s:%s", s.accessKeyID, digest)
	headers[HTTPHeaderAuthorization] = auth
	return nil
}

// add commonHeaders to headers after signature if not conflict
func addHeadersAfterSign(commonHeaders, headers map[string]string) {
	for k, v := range commonHeaders {
		lowerKey := strings.ToLower(k)
		if _, ok := headers[lowerKey]; !ok {
			headers[k] = v
		}
	}
}
