package bodycodegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
)

// Collections are immutable scalar sequences. Record copies can share their
// backing values because source bodies cannot assign elements or grow a slice.
type recordBodySequence struct {
	Kind   uint8
	Values []recordBodyScalar
}

func (s *recordBodySequence) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	values := make([]any, len(s.Values))
	for i, value := range s.Values {
		values[i] = recordBodyScalarValue(value)
	}
	return json.Marshal(values)
}

func decodeRecordSequence(raw []byte, sequence *types.Slice, records []RecordType) (*recordBodySequence, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		return nil, fmt.Errorf("repeated field requires an ordered array")
	}
	kind, optional, err := recordBodyScalarKind(sequence.Elem())
	if err != nil || optional || kind > 3 {
		return nil, fmt.Errorf("repeated field requires scalar elements")
	}
	var encoded []json.RawMessage
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil, err
	}
	if len(encoded) > RepeatedFieldLimit {
		return nil, fmt.Errorf("repeated field exceeds %d elements", RepeatedFieldLimit)
	}
	result := &recordBodySequence{Kind: kind, Values: make([]recordBodyScalar, len(encoded))}
	for i, raw := range encoded {
		value, err := decodeRecordCaseValue(raw, sequence.Elem(), records)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		result.Values[i], err = recordBodyScalarFromValue(value, kind)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
	}
	return result, nil
}

func equalRecordBodyScalar(a, b recordBodyScalar) bool {
	left, right := a.Collection, b.Collection
	a.Collection, b.Collection = nil, nil
	if a != b {
		return false
	}
	if left == nil || right == nil {
		return left == right
	}
	if left.Kind != right.Kind || len(left.Values) != len(right.Values) {
		return false
	}
	for i, value := range left.Values {
		if value != right.Values[i] {
			return false
		}
	}
	return true
}

func equalRecordBodyValue(a, b recordBodyValue) bool {
	if a.Type != b.Type || a.Count != b.Count {
		return false
	}
	for i := range a.Values {
		if !equalRecordBodyScalar(a.Values[i], b.Values[i]) {
			return false
		}
	}
	return true
}

// Scalar labels keep their existing representation; arrays use ordered JSON.
func recordScalarDisplay(value recordBodyScalar) string {
	if value.Kind >= 4 && value.Kind <= 6 {
		raw, err := json.Marshal(recordBodyScalarValue(value))
		if err == nil {
			return string(raw)
		}
	}
	return fmt.Sprint(recordScalarJSONValue(value))
}
