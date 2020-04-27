package filters

import (
	"k8s.io/api/core/v1"
	log "k8s.io/klog"
	"reflect"
	"regexp"
)

type GenericFilter struct {
	field  string
	keys   []string
	regexp bool
}

func (gf *GenericFilter) Filter(event *v1.Event) (matched bool) {
	if gf.keys == nil || len(gf.keys) == 0 {
		return false
	}

	field := reflect.Indirect(reflect.ValueOf(event)).FieldByName(gf.field)

	for _, k := range gf.keys {
		// enable regexp
		if gf.regexp {
			if ok, err := regexp.Match(k, []byte(field.String())); err == nil && ok {
				matched = true
				return
			} else {
				if err != nil {
					log.Errorf("Failed to match pattern %s with %s,because of %v", k, field.String(), err)
				}
				return false
			}
		} else {
			if field.String() == k {
				matched = true
				return
			}
		}
	}
	return false
}

// Generic Filter
func NewGenericFilter(field string, keys []string, regexp bool) *GenericFilter {
	k := &GenericFilter{
		field:  field,
		regexp: regexp,
	}
	if keys != nil {
		k.keys = keys
		return k
	}
	k.keys = make([]string, 0)
	return k
}
