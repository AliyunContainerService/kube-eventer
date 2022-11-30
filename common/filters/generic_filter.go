package filters

import (
	"reflect"
	"regexp"

	v1 "k8s.io/api/core/v1"
	log "k8s.io/klog"
)

type GenericFilter struct {
	field  string
	keys   []string
	regexp bool
}

func IsZero(v reflect.Value) bool {
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func (gf *GenericFilter) Filter(event *v1.Event) (matched bool) {
	var field reflect.Value

	switch gf.field {
	case "Kind":
		field = reflect.Indirect(reflect.ValueOf(event)).FieldByNameFunc(func(name string) bool {
			return name == "InvolvedObject"
		}).FieldByName("Kind")
	case "Namespace":
		field = reflect.Indirect(reflect.ValueOf(event)).FieldByNameFunc(func(name string) bool {
			return name == "InvolvedObject"
		}).FieldByName("Namespace")
	case "Type":
		field = reflect.Indirect(reflect.ValueOf(event)).FieldByName("Type")
	case "Reason":
		field = reflect.Indirect(reflect.ValueOf(event)).FieldByName("Reason")
	}

	if IsZero(field) {
		return false
	}

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
