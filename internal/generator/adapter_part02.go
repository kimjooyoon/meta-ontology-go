package generator

import (
	"fmt"
	"reflect"
)

func reflectedStruct(input any) (reflect.Value, error) {
	if input == nil {
		return reflect.Value{}, fmt.Errorf("generator: nil semantic input")
	}
	value := indirectValue(reflect.ValueOf(input))
	if !value.IsValid() {
		return reflect.Value{}, fmt.Errorf("generator: nil semantic input")
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("generator: unsupported semantic input %T", input)
	}
	return value, nil
}
func indirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}
func valueString(value reflect.Value) string {
	value = indirectValue(value)
	if !value.IsValid() {
		return ""
	}
	if value.Kind() == reflect.String {
		return value.String()
	}
	if value.CanInterface() {
		return fmt.Sprint(value.Interface())
	}
	return ""
}
