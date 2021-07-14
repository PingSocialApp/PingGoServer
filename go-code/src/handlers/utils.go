package handlers

import (
	"reflect"

	"github.com/microcosm-cc/bluemonday"
)

var p *bluemonday.Policy

func Init() {
	p = bluemonday.StrictPolicy()
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
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToDbMap(field)
			} else {
				res[tag] = field
			}
		}
	}
	return res
}
func structToJsonMap(item interface{}) map[string]interface{} {

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
		tag := v.Field(i).Tag.Get("jsonData")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToJsonMap(field)
			} else {
				if reflect.TypeOf(field).Kind() == reflect.String {
					res[tag] = (*p).Sanitize(field.(string))
				} else {
					res[tag] = field
				}
			}
		}
	}
	return res
}
