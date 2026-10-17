package bodycodegen

import (
	"fmt"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

func decodeRecordObjectCase(raw []byte, named *types.Named, records []RecordType) (any, error) {
	record := recordTypeByName(records, named.Obj().Name())
	if record == nil {
		return nil, fmt.Errorf("record type is not declared")
	}
	object, err := valuejson.Object(raw, RecordFieldLimit)
	if err != nil {
		return nil, err
	}
	required := 0
	for _, field := range record.Fields {
		if field.Presence != "optional" {
			required++
		}
	}
	if len(object) < required || len(object) > len(record.Fields) {
		return nil, fmt.Errorf("record requires exactly its declared fields")
	}
	declared := make(map[string]bool, len(record.Fields))
	for _, field := range record.Fields {
		declared[field.Name] = true
	}
	for name := range object {
		if !declared[name] {
			return nil, fmt.Errorf("record field %q is not declared", name)
		}
	}
	value, err := zeroRecordBodyValue(named)
	if err != nil {
		return nil, err
	}
	for i, field := range record.Fields {
		fieldRaw, ok := object[field.Name]
		if !ok {
			if field.Presence == "optional" {
				continue
			}
			return nil, fmt.Errorf("record field %q is missing", field.Name)
		}
		fieldType := types.Type(types.Typ[types.String])
		if field.TypeID == "urn:gooo:type:boolean" {
			fieldType = types.Typ[types.Bool]
		} else if field.TypeID == "urn:gooo:type:integer" {
			fieldType = types.Typ[types.Int64]
		}
		if field.Cardinality == "many" {
			fieldType = types.NewSlice(fieldType)
		}
		decoded, err := decodeRecordCaseField(fieldRaw, fieldType, records, field.TextBound)
		if err != nil {
			if field.TextBound.IsSet() {
				return nil, fmt.Errorf("record field %q (%s): %w", field.Name, field.ID, err)
			}
			return nil, err
		}
		kind, _, err := recordBodyScalarKind(fieldType)
		if err != nil {
			return nil, err
		}
		value.Values[i], err = recordBodyScalarFromValue(decoded, kind)
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}
