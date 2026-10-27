package transportorigins

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

// The native report's JSON tags own its physical shape. Decode only exact field
// names; unknown data stays in the original bytes, not in this decoding view.
// In particular, a case-folded alias cannot overwrite a canonical field.
func projectReportJSON(raw []byte, shape reflect.Type) (json.RawMessage, bool, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return raw, shape.Kind() == reflect.Pointer || shape.Kind() == reflect.Slice, nil
	}
	if shape.Kind() == reflect.Pointer {
		return projectReportJSON(raw, shape.Elem())
	}
	if shape.Kind() != reflect.Struct {
		return raw, true, nil
	}
	fields, err := valuejson.Object(raw, len(raw))
	if err != nil {
		return nil, false, err
	}
	known := make(map[string]json.RawMessage, shape.NumField())
	complete := true
	for index := range shape.NumField() {
		field := shape.Field(index)
		if field.PkgPath != "" {
			continue
		}
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		value, present := fields[name]
		if !present {
			complete = complete && strings.Contains(","+options+",", ",omitempty,")
			continue
		}
		projected, available, err := projectReportJSON(value, field.Type)
		if err != nil {
			return nil, false, err
		}
		known[name], complete = projected, complete && available
	}
	projected, err := json.Marshal(known)
	return projected, complete, err
}
