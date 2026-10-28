package bodycodegen

import (
	"fmt"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

func decodeRecordObjectCase(raw []byte, named *types.Named, records []RecordType) (any, error) {
	decoder := newRecordCaseDecoder(named, records)
	return decoder.object(raw, named)
}

func (d recordCaseDecoder) object(raw []byte, named *types.Named) (any, error) {
	record := recordTypeByName(d.records, named.Obj().Name())
	if record == nil {
		return nil, fmt.Errorf("record type is not declared")
	}
	structure, ok := named.Underlying().(*types.Struct)
	if !ok || structure.NumFields() != len(record.Fields) {
		return nil, fmt.Errorf("record layout differs from its declaration")
	}
	object, err := valuejson.Object(raw, RecordFieldLimit)
	if err != nil {
		return nil, err
	}
	if err := d.charge(len(object) + 1); err != nil {
		return nil, err
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
		if structure.Field(i).Name() != field.Name && structure.Field(i).Name() != field.GoName {
			return nil, fmt.Errorf("record field order differs from its declaration")
		}
		fieldType, err := recordCaseFieldType(field, structure.Field(i).Type(), d.records)
		if err != nil {
			return nil, err
		}
		fieldRaw, exists := object[field.Name]
		if !exists {
			if field.Presence == "optional" {
				continue
			}
			return nil, fmt.Errorf("record field %q is missing", field.Name)
		}
		value.Values[i], err = d.objectField(fieldRaw, fieldType, field)
		if err != nil {
			return nil, err
		}

	}
	return value, validateRecordBodyValueBudget(value)
}

func recordCaseFieldType(field RecordField, t types.Type, records []RecordType) (types.Type, error) {
	original := t
	if pointer, optional := t.(*types.Pointer); optional {
		if field.Presence != "optional" {
			return nil, fmt.Errorf("record field %q has a different presence contract", field.ID)
		}
		t = pointer.Elem()
	} else if field.Presence == "optional" {
		return nil, fmt.Errorf("record field %q requires an optional type", field.ID)
	}
	decoded := t
	if sequence, many := t.(*types.Slice); many {
		if field.Cardinality != "many" {
			return nil, fmt.Errorf("record field %q has a different collection contract", field.ID)
		}
		t = sequence.Elem()
	} else if field.Cardinality == "many" {
		return nil, fmt.Errorf("record field %q requires a collection type", field.ID)
	}
	expected := map[string]types.Type{
		"urn:gooo:type:string":  types.Typ[types.String],
		"urn:gooo:type:boolean": types.Typ[types.Bool],
		"urn:gooo:type:integer": types.Typ[types.Int64],
	}
	if primitive := expected[field.TypeID]; primitive != nil {
		if t == primitive {
			return decoded, nil
		}
	} else {
		for _, target := range records {
			if name := recordBodyElementName(t); name != "" && target.ID == field.TypeID && (name == target.Name || name == target.GoName) {
				return decoded, nil
			}
		}
	}
	return nil, fmt.Errorf("record field %q type %q differs from %s", field.ID, field.TypeID, original)
}

func (d recordCaseDecoder) objectField(raw []byte, t types.Type, field RecordField) (recordBodyScalar, error) {
	decoded, err := d.field(raw, t, field.TextBound)
	if err != nil {
		return recordBodyScalar{}, fmt.Errorf("record field %q (%s): %w", field.Name, field.ID, err)
	}
	kind, _, err := recordBodyScalarKind(t)
	if err != nil {
		return recordBodyScalar{}, err
	}
	value, err := recordBodyScalarFromValue(decoded, kind)
	if err != nil {
		return recordBodyScalar{}, fmt.Errorf("record field %q (%s): %w", field.Name, field.ID, err)
	}
	return value, nil
}
