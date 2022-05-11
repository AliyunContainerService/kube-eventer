package metrichub

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func map2QueryParams(m http.Header) string {
	const capacity = 3
	keySlice := make([]string, 0, capacity)
	for k, v := range m {
		if len(v) > 0 && (strings.HasPrefix(k, XAcs) || strings.HasPrefix(k, XCms)) {
			keySlice = append(keySlice, k+":"+strings.TrimSpace(v[0]))
		}
	}

	sort.Strings(keySlice)

	return strings.Join(keySlice, "\n")
}

func sortQuery(query url.Values) (r string) {
	if len(query) > 0 {
		keys := make([]string, 0, len(query))
		for k := range query {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		items := make([]string, 0, len(keys))
		for _, k := range keys {
			for _, v := range query[k] {
				items = append(items, k+"="+v)
			}
		}
		r = "?" + strings.Join(items, "&")
	}
	return
}

// Signature calculates a request's signature digest.
func (p *Client) Signature(method, uri string, headers http.Header) (digest, signStr string, err error) {
	date := headers.Get(Date)
	if date == "" {
		return "", "", fmt.Errorf("can't find 'Date' header")
	}

	var u *url.URL
	if u, err = url.Parse(uri); err == nil {
		get := func(key string) (r string) {
			if v := headers[key]; len(v) > 0 {
				r = v[0]
			}
			return
		}
		signStr = method + "\n" +
			get(ContentMd5) + "\n" +
			get(ContentType) + "\n" +
			date + "\n" +
			map2QueryParams(headers) + "\n" +
			u.EscapedPath() + sortQuery(u.Query())

		mac := hmac.New(sha1.New, []byte(p.accessSecret))
		if _, err = mac.Write([]byte(signStr)); err == nil {
			digest = strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
		}
	}
	return
}

func (p *Client) SignatureRequest(request *http.Request) (digest, signStr string, err error) {
	return p.Signature(request.Method, request.URL.RequestURI(), request.Header)
}
