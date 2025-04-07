package sls

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func BoolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func BoolPtrToStringNum(b *bool) string {
	if b == nil {
		return ""
	}
	if *b {
		return "1"
	}
	return "0"
}

func Int64PtrToString(i *int64) string {
	if i == nil {
		return ""
	}
	return strconv.FormatInt(*i, 10)
}

func ParseHeaderInt(r *http.Response, headerName string) (int, error) {
	values := r.Header[headerName]
	if len(values) > 0 {
		value, err := strconv.Atoi(values[0])
		if err != nil {
			return -1, fmt.Errorf("can't parse '%s' header: %v", strings.ToLower(headerName), err)
		}
		return value, nil
	}
	return -1, fmt.Errorf("can't find '%s' header", strings.ToLower(headerName))
}
