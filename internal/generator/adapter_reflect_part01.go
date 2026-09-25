package generator

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type reflectedCollection struct {
	values  []reflect.Value
	ordered bool
}

func readReflectedCollection(value reflect.Value, name string, input any) (reflectedCollection, error) {
	field := value.FieldByName(name)
	if !field.IsValid() {
		return reflectedCollection{}, fmt.Errorf("generator: semantic input %T does not expose %s; implement SemanticIRProvider", input, name)
	}
	field = indirectValue(field)
	if !field.IsValid() {
		return reflectedCollection{}, fmt.Errorf("generator: semantic input %T has nil %s", input, name)
	}
	switch field.Kind() {
	case reflect.Array, reflect.Slice:
		values := make([]reflect.Value, field.Len())
		for index := range values {
			values[index] = field.Index(index)
		}
		return reflectedCollection{values: values, ordered: true}, nil
	case reflect.Map:
		keys := field.MapKeys()
		sort.SliceStable(keys, func(left, right int) bool {
			return reflectedKey(keys[left]) < reflectedKey(keys[right])
		})
		values := make([]reflect.Value, 0, len(keys))
		for _, key := range keys {
			values = append(values, field.MapIndex(key))
		}
		return reflectedCollection{values: values}, nil
	default:
		return reflectedCollection{}, fmt.Errorf("generator: semantic input %T has unsupported %s collection %s", input, name, field.Kind())
	}
}
func reflectedKey(value reflect.Value) string {
	return value.Kind().String() + ":" + valueString(value)
}
func optionalStringField(value reflect.Value, name string) (string, error) {
	field := value.FieldByName(name)
	if !field.IsValid() {
		return "", nil
	}
	field = indirectValue(field)
	if !field.IsValid() || field.Kind() != reflect.String {
		return "", fmt.Errorf("generator: semantic field %s must be a string", name)
	}
	return strings.TrimSpace(field.String()), nil
}
func requiredStringField(value reflect.Value, name, context string) (string, error) {
	field := value.FieldByName(name)
	if !field.IsValid() {
		return "", fmt.Errorf("generator: %s is missing %s", context, name)
	}
	field = indirectValue(field)
	if !field.IsValid() || field.Kind() != reflect.String {
		return "", fmt.Errorf("generator: %s field %s must be a string", context, name)
	}
	result := strings.TrimSpace(field.String())
	if result == "" {
		return "", fmt.Errorf("generator: %s has empty %s", context, name)
	}
	return result, nil
}
