package bodycodegen

import (
	"fmt"
	"go/types"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

func decodeRecordCaseField(raw []byte, t types.Type, records []RecordType, bound textbound.UTF8) (any, error) {
	if err := bound.Validate(); err != nil {
		return nil, err
	}
	if !bound.IsSet() {
		return decodeRecordCaseValue(raw, t, records)
	}
	if sequence, ok := t.(*types.Slice); ok {
		return decodeRecordSequenceWithBound(raw, sequence, records, bound)
	}
	maximum, err := bound.Resolve(0)
	if err != nil || t != types.Typ[types.String] {
		return nil, fmt.Errorf("max_utf8_bytes requires a Text field")
	}
	return valuejson.DecodeText(raw, maximum)
}

func validateRecordOutputBounds(value recordBodyValue, record RecordType) error {
	for i, field := range record.Fields {
		if !field.TextBound.IsSet() || field.Presence == "optional" && !value.Values[i].Present {
			continue
		}
		maximum, err := field.TextBound.Resolve(0)
		if err != nil || field.TypeID != "urn:gooo:type:string" {
			return fmt.Errorf("record field %q (%s) has an invalid Text byte bound", field.Name, field.ID)
		}
		values := value.Values[i : i+1]
		if field.Cardinality == "many" {
			if value.Values[i].Collection == nil {
				return fmt.Errorf("record field %q (%s) requires a collection", field.Name, field.ID)
			}
			values = value.Values[i].Collection.Values
		}
		for _, scalar := range values {
			if scalar.Kind != 1 || !utf8.ValidString(scalar.Text) {
				return fmt.Errorf("record field %q (%s): Text requires valid UTF-8", field.Name, field.ID)
			}
			if uint64(len(scalar.Text)) > uint64(maximum) {
				return fmt.Errorf("record field %q (%s): Text exceeds %d UTF-8 bytes", field.Name, field.ID, maximum)
			}
		}
	}
	return nil
}

func validateNestedRecordOutput(value recordBodyValue, record RecordType, records []RecordType) error {
	if err := validateRecordBodyValueBudget(value); err != nil {
		return err
	}
	return validateNestedRecordOutputAt(value, record, records, 1)
}

func validateNestedRecordOutputAt(value recordBodyValue, record RecordType, records []RecordType, depth int) error {
	if depth > syntax.RecordReferenceDepthLimit || (value.Type != record.Name && value.Type != record.GoName) || value.Count != len(record.Fields) {
		return fmt.Errorf("record output differs from its nominal layout")
	}
	if err := validateRecordOutputBounds(value, record); err != nil {
		return err
	}
	for i, field := range record.Fields {
		scalar := value.Values[i]
		if !scalar.Present {
			continue
		}
		children := []*recordBodyValue{scalar.Record}
		if scalar.Collection != nil {
			children = make([]*recordBodyValue, len(scalar.Collection.Values))
			for j, item := range scalar.Collection.Values {
				children[j] = item.Record
			}
		}
		for _, child := range children {
			if child == nil {
				continue
			}
			target := recordTypeByName(records, child.Type)
			if target == nil || target.ID != field.TypeID {
				return fmt.Errorf("record field %q has a different nominal type", field.ID)
			}
			if err := validateNestedRecordOutputAt(*child, *target, records, depth+1); err != nil {
				return fmt.Errorf("record field %q (%s): %w", field.Name, field.ID, err)
			}
		}
	}
	return nil
}
