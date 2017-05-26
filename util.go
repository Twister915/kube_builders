package kubernetes

import (
	"fmt"
	"reflect"
)

func setAtMap(target *map[string]string, key string, value interface{}) {
	if *target == nil {
		*target = make(map[string]string)
	}
	(*target)[key] = fmt.Sprintf("%v", value)
}

func setAtMapDirect(target, key, value interface{}) {
	rTarget := reflect.ValueOf(target)
	if rTarget.Elem().IsNil() {
		rTarget.Elem().Set(reflect.MakeMap(rTarget.Type().Elem()))
	}
	rTarget.Elem().SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
}