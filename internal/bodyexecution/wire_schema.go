package bodyexecution

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// A decoder invocation owns this shape table. Source records are never cached.
// Exact field names and duplicate checks share one traversal of the wire bytes.
type wireSchema struct {
	fields map[reflect.Type]map[string]reflect.Type
}

func (schema *wireSchema) scanValue(d *json.Decoder, depth int, typ reflect.Type) error {
	if depth > 64 {
		return fmt.Errorf("JSON exceeds 64 nested levels")
	}
	typ = wireType(typ)
	value, err := d.Token()
	if err != nil {
		return err
	}
	if typ == reflect.TypeFor[[]byte]() && value != nil {
		if _, ok := value.(string); !ok {
			return fmt.Errorf("JSON byte slices are base64 strings")
		}
	}
	delim, ok := value.(json.Delim)
	if !ok {
		return nil
	}
	return schema.scanContainer(d, depth, typ, delim)
}

func (schema *wireSchema) scanContainer(d *json.Decoder, depth int, typ reflect.Type, delim json.Delim) error {
	keys := map[string]bool{}
	for d.More() {
		child := wireElement(typ)
		if delim == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || keys[name] {
				return fmt.Errorf("duplicate or invalid JSON key")
			}
			keys[name] = true
			child, err = schema.field(typ, name)
			if err != nil {
				return err
			}
		}
		if err := schema.scanValue(d, depth+1, child); err != nil {
			return err
		}
	}
	_, err := d.Token()
	return err
}

func (schema *wireSchema) field(typ reflect.Type, name string) (reflect.Type, error) {
	if typ == nil || typ.Kind() != reflect.Struct {
		return wireElement(typ), nil
	}
	fields, ok := schema.fields[typ]
	if !ok {
		fields = map[string]reflect.Type{}
		for field := range typ.Fields() {
			key, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			if key != "" && key != "-" {
				fields[key] = field.Type
			}
		}
		schema.fields[typ] = fields
	}
	child, ok := fields[name]
	if !ok {
		return nil, fmt.Errorf("unknown or aliased JSON field %q", name)
	}
	return child, nil
}

func wireType(typ reflect.Type) reflect.Type {
	for typ != nil && typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflect.TypeFor[json.RawMessage]() {
		return nil
	}
	return typ
}

func wireElement(typ reflect.Type) reflect.Type {
	if typ != nil {
		switch typ.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return typ.Elem()
		}
	}
	return nil
}
