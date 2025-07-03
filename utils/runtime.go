package utils

import (
	"reflect"
)

func IsNilInterface(i interface{}) bool {
	if i == nil {
		return true
	}
	val := reflect.ValueOf(i)
	return val.Kind() == reflect.Ptr && val.IsNil()
}
