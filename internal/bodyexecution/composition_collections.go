package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func canonicalRecordField(raw []byte, scalar, cardinality string) (json.RawMessage, error) {
	return canonicalRecordFieldWithBound(raw, scalar, cardinality, 0)
}

func canonicalRecordFieldWithBound(raw []byte, scalar, cardinality string, bound textbound.UTF8) (json.RawMessage, error) {
	if err := bound.Validate(); err != nil {
		return nil, err
	}
	if bound.IsSet() && scalar != "Text" {
		return nil, fmt.Errorf("max_utf8_bytes requires a Text field")
	}
	if cardinality != "many" {
		return canonicalFieldScalar(raw, scalar, bound)
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		return nil, fmt.Errorf("repeated field requires an array; null is not a collection")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	if len(values) > bodycodegen.RepeatedFieldLimit {
		return nil, fmt.Errorf("repeated field has %d values, limit %d", len(values), bodycodegen.RepeatedFieldLimit)
	}
	for i, value := range values {
		canonical, err := canonicalFieldScalar(value, scalar, bound)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		values[i] = canonical
	}
	return json.Marshal(values)
}

func canonicalFieldScalar(raw []byte, scalar string, bound textbound.UTF8) (json.RawMessage, error) {
	if !bound.IsSet() {
		return canonicalScalar(raw, scalar)
	}
	if scalar != "Text" {
		return nil, fmt.Errorf("max_utf8_bytes requires a Text field")
	}
	text, err := decodeBoundedTextScalar(raw, bound)
	if err != nil {
		return nil, err
	}
	return json.Marshal(text)
}
