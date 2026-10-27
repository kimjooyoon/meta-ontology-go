package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func canonicalNestedField(raw []byte, field bodycodegen.RecordField, records []bodycodegen.RecordType, depth int, slots *int) (json.RawMessage, error) {
	if scalar, ok := compositionRecordFieldScalar(field.TypeID); ok {
		canonical, err := canonicalRecordFieldWithBound(raw, scalar, field.Cardinality, field.TextBound)
		if err != nil {
			return nil, err
		}
		if *slots >= 0 && field.Cardinality == "many" {
			var values []json.RawMessage
			if err := json.Unmarshal(canonical, &values); err != nil {
				return nil, err
			}
			*slots += len(values)
			if *slots > syntax.RecordValueSlotLimit {
				return nil, fmt.Errorf("record value exceeds %d slots", syntax.RecordValueSlotLimit)
			}
		}
		return canonical, nil
	}
	var target *bodycodegen.RecordType
	for i := range records {
		if records[i].ID == field.TypeID {
			target = &records[i]
			break
		}
	}
	if target == nil || field.TextBound.IsSet() {
		return nil, fmt.Errorf("unsupported record type %q", field.TypeID)
	}
	if field.Cardinality != "many" {
		return canonicalNestedRecord(raw, *target, records, depth+1, slots)
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		return nil, fmt.Errorf("record collection requires an array")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	if len(values) > bodycodegen.RepeatedFieldLimit {
		return nil, fmt.Errorf("record collection exceeds %d elements", bodycodegen.RepeatedFieldLimit)
	}
	for i, value := range values {
		canonical, err := canonicalNestedRecord(value, *target, records, depth+1, slots)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		values[i] = canonical
	}
	return json.Marshal(values)
}

// Nested layouts share a decoded JSON node budget. Independent scalar profiles
// retain their existing limits; this bound applies to the newly nested value.
func chargeNestedRecordSlots(values map[string]json.RawMessage, slots *int) error {
	if *slots < 0 {
		return nil
	}
	*slots += len(values) + 1
	if *slots > syntax.RecordValueSlotLimit {
		return fmt.Errorf("record value exceeds %d slots", syntax.RecordValueSlotLimit)
	}
	return nil
}

// This is a decoded value/source-layout projection. It does not establish
// dynamic dataflow, invocation identity or authentication for nested values.
func (graph compositionGraph) nestedFieldValues(field bodycodegen.RecordField, raw []byte) []CompositionRecordField {
	for _, target := range graph.plan.Records {
		if target.ID != field.TypeID {
			continue
		}
		if field.Cardinality != "many" {
			return graph.recordFieldValues(target.Name, raw)
		}
		var values []json.RawMessage
		if json.Unmarshal(raw, &values) != nil {
			return nil
		}
		result := make([]CompositionRecordField, len(values))
		for i, value := range values {
			result[i] = CompositionRecordField{ID: target.ID, Name: target.Name, Value: value, Index: new(i), Children: graph.recordFieldValues(target.Name, value)}
		}
		return result
	}
	return nil
}
