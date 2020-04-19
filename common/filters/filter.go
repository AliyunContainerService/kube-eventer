package filters

import (
	"k8s.io/api/core/v1"
	"strings"
)

// All filter interface
type Filter interface {
	Filter(event *v1.Event) (matched bool)
}

func GetValues(o []string) []string {
	if len(o) >= 1 {
		if len(o[0]) == 0 {
			return nil
		}
		return strings.Split(o[0], ",")
	}
	return nil
}
