package util

import (
	"fmt"
	"regexp"
	"strings"
)

const ENDPOINT_REGEX_PATTERN = `^(?:http[s]?:\/\/)?([a-z-0-9]+)\.(?:sls|log)\.aliyuncs\.com$`

var regionSuffixs = []string{"-intranet", "-share", "-vpc"}

func ParseRegion(endpoint string) (string, error) {
	var re = regexp.MustCompile(ENDPOINT_REGEX_PATTERN)
	groups := re.FindStringSubmatch(endpoint)
	if groups == nil {
		return "", fmt.Errorf("invalid endpoint format: %s", endpoint)
	}
	region := groups[1]
	for _, suffix := range regionSuffixs {
		if strings.HasSuffix(region, suffix) {
			return region[:len(region)-len(suffix)], nil
		}
	}
	return region, nil
}
