// This file is auto-generated, don't edit it. Thanks.
/**
 * This is for EventBridge SDK
 */
package client

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"hash"
	"io"
	"sort"
	"strings"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Sorter struct {
	Keys []string
	Vals []string
}

func newSorter(m map[string]string) *Sorter {
	hs := &Sorter{
		Keys: make([]string, 0, len(m)),
		Vals: make([]string, 0, len(m)),
	}

	for k, v := range m {
		hs.Keys = append(hs.Keys, k)
		hs.Vals = append(hs.Vals, v)
	}
	return hs
}

// Sort is an additional function for function SignHeader.
func (hs *Sorter) Sort() {
	sort.Sort(hs)
}

// Len is an additional function for function SignHeader.
func (hs *Sorter) Len() int {
	return len(hs.Vals)
}

// Less is an additional function for function SignHeader.
func (hs *Sorter) Less(i, j int) bool {
	return bytes.Compare([]byte(hs.Keys[i]), []byte(hs.Keys[j])) < 0
}

// Swap is an additional function for function SignHeader.
func (hs *Sorter) Swap(i, j int) {
	hs.Vals[i], hs.Vals[j] = hs.Vals[j], hs.Vals[i]
	hs.Keys[i], hs.Keys[j] = hs.Keys[j], hs.Keys[i]
}

/**
 * Get the string to be signed according to request
 * @param request  which contains signed messages
 * @return the signed string
 */
func GetStringToSign(request *tea.Request) (_result *string) {
	return tea.String(getStringToSign(request))
}

func getStringToSign(request *tea.Request) string {
	resource := tea.StringValue(request.Pathname)
	queryParams := request.Query
	// sort QueryParams by key
	var queryKeys []string
	for key := range queryParams {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)
	tmp := ""
	for i := 0; i < len(queryKeys); i++ {
		queryKey := queryKeys[i]
		tmp = tmp + "&" + queryKey + "=" + tea.StringValue(queryParams[queryKey])
	}
	if tmp != "" {
		tmp = strings.TrimLeft(tmp, "&")
		resource = resource + "?" + tmp
	}
	return getSignedStr(request, resource)
}

func getSignedStr(req *tea.Request, canonicalizedResource string) string {
	temp := make(map[string]string)

	for k, v := range req.Headers {
		if strings.HasPrefix(strings.ToLower(k), "x-acs") {
			temp[strings.ToLower(k)] = tea.StringValue(v)
		}
	}
	hs := newSorter(temp)

	// Sort the temp by the ascending order
	hs.Sort()

	// Get the canonicalizedOSSHeaders
	canonicalizedHeaders := ""
	for i := range hs.Keys {
		canonicalizedHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
	}

	// Give other parameters values
	// when sign URL, date is expires
	date := tea.StringValue(req.Headers["date"])
	contentType := tea.StringValue(req.Headers["content-type"])
	contentMd5 := tea.StringValue(req.Headers["content-md5"])

	signStr := tea.StringValue(req.Method) + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedHeaders + canonicalizedResource
	return signStr
}

/**
 * Get signature according to stringToSign, secret
 * @param stringToSign  the signed string
 * @param secret accesskey secret
 * @return the signature
 */
func GetSignature(stringToSign *string, secret *string) (_result *string) {
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(tea.StringValue(secret)))
	io.WriteString(h, tea.StringValue(stringToSign))
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return tea.String(signedStr)
}

/**
 * Encode data in events
 * @param events the object
 * @return the result
 */
func Serialize(events interface{}) (_result interface{}) {
	if tea.BoolValue(util.IsUnset(events)) {
		return events
	}
	out := make([]interface{}, 0)
	byt, _ := json.Marshal(events)
	err := json.Unmarshal(byt, &out)
	if err != nil {
		return events
	}

	for i := 0; i < len(out); i++ {
		tmp := out[i]
		m := make(map[string]interface{})
		byt, _ = json.Marshal(tmp)
		json.Unmarshal(byt, &m)

		if m["datacontenttype"] != nil {
			datacontenttype := m["datacontenttype"].(string)
			if (!strings.HasPrefix(datacontenttype, "application/json") &&
				!strings.HasPrefix(datacontenttype, "text/json")) && m["data"] != nil {
				data := m["data"].(string)
				m["data_base64"] = data
				delete(m, "data")
			}
		}

		if m["data"] != nil {
			var res interface{}
			data := m["data"].(string)
			tmp, _ := base64.StdEncoding.DecodeString(data)
			err = json.Unmarshal(tmp, &res)
			if err != nil {
				m["data"] = string(tmp)
			} else {
				m["data"] = res
			}
		}

		if m["extensions"] != nil {
			extensions := m["extensions"].(map[string]interface{})
			for k, v := range extensions {
				m[k] = v
			}
			delete(m, "extensions")
		}
		out[i] = m
	}

	return out
}

/**
 * Judge if the  origin is start with the prefix
 * @param origin the original string
 * @param prefix the prefix string
 * @return the result
 */
func StartWith(origin, prefix *string) (_result *bool) {
	res := strings.HasPrefix(tea.StringValue(origin), tea.StringValue(prefix))
	return tea.Bool(res)
}
