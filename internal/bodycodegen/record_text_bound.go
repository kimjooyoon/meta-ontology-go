package bodycodegen

import (
	"fmt"
	"go/types"
	"unicode/utf8"

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
