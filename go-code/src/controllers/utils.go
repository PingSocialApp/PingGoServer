package controllers

import (
	"html"
	"reflect"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

var sanitizerPolicy *bluemonday.Policy

func Init() {
	sanitizerPolicy = bluemonday.UGCPolicy()
}

func structToDbMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}

	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	if reflectValue.IsZero() {
		return res
	}
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("db")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if t, ok := field.(time.Time); ok {
				res[tag] = t.UTC().Format(time.RFC3339)
			} else if v.Field(i).Type.Kind() == reflect.Ptr || v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToDbMap(field)
			} else {
				if reflect.TypeOf(field).Kind() == reflect.String {
					res[tag] = html.UnescapeString(sanitizerPolicy.Sanitize(field.(string)))
				} else {
					res[tag] = field
				}
			}
		}
	}
	return res
}

func ValueExtractor(data interface{}, exists bool) (ret interface{}) {
	if exists {
		return data
	} else {
		return nil
	}
}
func isNilFixed(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
