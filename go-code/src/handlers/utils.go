package handlers

import (
	"reflect"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var sanitizerPolicy *bluemonday.Policy

func Init() {
	sanitizerPolicy = bluemonday.StrictPolicy()
}

func structToDbMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("db")

		// remove omitEmpty
		omitEmpty := false
		if strings.HasSuffix(tag, "omitempty") {
			omitEmpty = true
			idx := strings.Index(tag, ",")
			if idx > 0 {
				tag = tag[:idx]
			} else {
				tag = ""
			}
		}

		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToDbMap(field)
			} else {
				if !(omitEmpty && reflectValue.Field(i).IsZero()) {
					if reflect.TypeOf(field).Kind() == reflect.String {
						res[tag] = sanitizerPolicy.Sanitize(field.(string))
					} else {
						res[tag] = field
					}
				}
			}
		}
	}
	return res
}
