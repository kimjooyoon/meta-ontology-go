package completeness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// checkShape derives required JSON members from the generated Go boundary,
// avoiding a second handwritten field list. Optional-one is explicitly nullable;
// required-many and required records must be arrays and objects, not null.
func checkShape(data json.RawMessage, typ reflect.Type) error {
	if typ.Kind() == reflect.Pointer {
		if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
			return nil
		}
		return checkShape(data, typ.Elem())
	}
	switch typ.Kind() {
	case reflect.Struct:
		var object map[string]json.RawMessage
		if err := json.Unmarshal(data, &object); err != nil {
			return err
		}
		if object == nil {
			return fmt.Errorf("required receipt object is null")
		}
		if len(object) != typ.NumField() {
			return fmt.Errorf("receipt object contains unknown or missing fields")
		}
		for field := range typ.Fields() {
			name := field.Tag.Get("json")
			value, ok := object[name]
			if !ok {
				return fmt.Errorf("required receipt field %q is absent", name)
			}
			if err := checkShape(value, field.Type); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	case reflect.Slice:
		var list []json.RawMessage
		if err := json.Unmarshal(data, &list); err != nil {
			return err
		}
		if list == nil {
			return fmt.Errorf("required receipt array is null")
		}
		for _, item := range list {
			if err := checkShape(item, typ.Elem()); err != nil {
				return err
			}
		}
	case reflect.Map:
		var object map[string]json.RawMessage
		if err := json.Unmarshal(data, &object); err != nil {
			return err
		}
		if object == nil {
			return fmt.Errorf("required receipt record is null")
		}
		for _, item := range object {
			if err := checkShape(item, typ.Elem()); err != nil {
				return err
			}
		}
	case reflect.Interface:
		return nil // source-owned opaque scope values; no completeness inference
	default:
		if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
			return fmt.Errorf("required scalar is null")
		}
	}
	return nil
}
