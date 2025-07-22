package ecsmetadata

import (
	"encoding/json"
	"strings"
)

func parsePathNames(data string) []string {
	var ret []string
	parts := strings.Split(data, "\n")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.TrimRight(part, "/")
		if part == "" {
			continue
		}
		ret = append(ret, part)
	}
	return ret
}

func parseJSONStringArray(data []byte) ([]string, error) {
	var ret []string
	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
