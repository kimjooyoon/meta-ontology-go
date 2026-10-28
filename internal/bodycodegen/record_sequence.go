package bodycodegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

// Collections are immutable scalar sequences. Record copies can share their
// backing values because source bodies cannot assign elements or grow a slice.
type recordBodySequence struct {
	Kind    uint8
	Element string
	Values  []recordBodyScalar
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
	return decodeRecordSequenceWithBound(raw, sequence, records, 0)
}

func decodeRecordSequenceWithBound(raw []byte, sequence *types.Slice, records []RecordType, bound textbound.UTF8) (*recordBodySequence, error) {
	decoder := newRecordCaseDecoder(sequence, records)
	return decoder.sequence(raw, sequence, bound)
}

func (d recordCaseDecoder) sequence(raw []byte, sequence *types.Slice, bound textbound.UTF8) (*recordBodySequence, error) {
	if bound.Validate() != nil || bound.IsSet() && sequence.Elem() != types.Typ[types.String] {
		return nil, fmt.Errorf("max_utf8_bytes requires Text collection elements")
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		return nil, fmt.Errorf("repeated field requires an ordered array")
	}
	kind, optional, err := recordBodyScalarKind(sequence.Elem())
	if err != nil || optional || !recordBodyElementKind(kind) {
		return nil, fmt.Errorf("repeated field requires declared scalar or record elements")
	}
	var encoded []json.RawMessage
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil, err
	}
	if len(encoded) > RepeatedFieldLimit {
		return nil, fmt.Errorf("repeated field exceeds %d elements", RepeatedFieldLimit)
	}
	if kind != recordBodyRecordKind {
		if err := d.charge(len(encoded)); err != nil {
			return nil, err
		}
	}
	result := &recordBodySequence{Kind: kind, Element: recordBodyElementName(sequence.Elem()), Values: make([]recordBodyScalar, len(encoded))}
	for i, raw := range encoded {
		value, err := d.field(raw, sequence.Elem(), bound)
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
	first, second := a.Record, b.Record
	a.Collection, b.Collection, a.Record, b.Record = nil, nil, nil, nil
	if a != b {
		return false
	}
	if first != nil || second != nil {
		if first == nil || second == nil || !equalRecordBodyValue(*first, *second) {
			return false
		}
	}
	if left == nil || right == nil {
		return left == right
	}
	if left.Kind != right.Kind || left.Element != right.Element || len(left.Values) != len(right.Values) {
		return false
	}
	for i, value := range left.Values {
		if !equalRecordBodyScalar(value, right.Values[i]) {
			return false
		}
	}
	return true
}

func equalRecordBodyValue(a, b recordBodyValue) bool {
	if a.Type != b.Type || a.Count != b.Count || a.Count < 1 || a.Count > RecordFieldLimit {
		return false
	}
	if validateRecordBodyValueBudget(a) != nil || validateRecordBodyValueBudget(b) != nil {
		return false
	}
	for i := 0; i < a.Count; i++ {
		if !equalRecordBodyScalar(a.Values[i], b.Values[i]) {
			return false
		}
	}
	return true
}

// Scalar labels keep their existing representation; arrays use ordered JSON.
func recordScalarDisplay(value recordBodyScalar) string {
	if recordBodyCollectionKind(value.Kind) || value.Kind == recordBodyRecordKind {
		raw, err := json.Marshal(recordBodyScalarValue(value))
		if err == nil {
			return string(raw)
		}
	}
	return fmt.Sprint(recordScalarJSONValue(value))
}
