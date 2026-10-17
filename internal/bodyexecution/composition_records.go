package bodyexecution

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

type CompositionRecordField struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Value    json.RawMessage `json:"value"`
	Presence string          `json:"presence,omitempty"`
	Present  *bool           `json:"present,omitempty"`
}

func (graph compositionGraph) recordType(name string) (bodycodegen.RecordType, bool) {
	for _, record := range graph.plan.Records {
		if record.Name == name {
			return record, true
		}
	}
	return bodycodegen.RecordType{}, false
}

func (graph compositionGraph) valueGoType(name string) string {
	if scalar := scalarGoType(graph.scalarKind(name)); scalar != "" {
		return scalar
	}
	if record, ok := graph.recordType(name); ok {
		return record.GoName
	}
	return ""
}

func (graph compositionGraph) canonicalValue(raw []byte, name string) (json.RawMessage, error) {
	if record, ok := graph.recordType(name); ok {
		return canonicalRecord(raw, record)
	}
	return canonicalScalar(raw, graph.scalarKind(name))
}

func canonicalRecord(raw []byte, record bodycodegen.RecordType) (json.RawMessage, error) {
	values, err := recordObject(raw)
	if err != nil {
		return nil, fmt.Errorf("record %q: %w", record.Name, err)
	}
	canonicalValues := make(map[string]json.RawMessage, len(values))
	for _, field := range record.Fields {
		value, present := values[field.Name]
		if !present {
			if field.Presence == "optional" {
				continue
			}
			return nil, fmt.Errorf("record %q requires field %q", record.Name, field.Name)
		}
		fieldType, ok := compositionRecordFieldScalar(field.TypeID)
		if !ok {
			return nil, fmt.Errorf("record %q field %q has unsupported type %q", record.Name, field.Name, field.TypeID)
		}
		canonical, err := canonicalRecordFieldWithBound(value, fieldType, field.Cardinality, field.TextBound)
		if err != nil {
			if field.TextBound.IsSet() {
				return nil, fmt.Errorf("record %q field %q (%s): %w", record.Name, field.Name, field.ID, err)
			}
			return nil, fmt.Errorf("record %q field %q: %w", record.Name, field.Name, err)
		}
		canonicalValues[field.Name] = canonical
		delete(values, field.Name)
	}
	if len(values) != 0 {
		return nil, fmt.Errorf("record %q contains an undeclared field", record.Name)
	}
	return json.Marshal(canonicalValues)
}

func compositionRecordFieldScalar(typeID string) (string, bool) {
	switch typeID {
	case string(semantic.BuiltinStringTypeID):
		return "Text", true
	case string(semantic.BuiltinBooleanTypeID):
		return "Boolean", true
	case string(semantic.BuiltinIntegerTypeID):
		return "Integer", true
	default:
		return "", false
	}
}

func recordObject(raw []byte) (map[string]json.RawMessage, error) {
	return valuejson.Object(raw, bodycodegen.RecordFieldLimit)
}

func (graph compositionGraph) recordFieldValues(name string, raw []byte) []CompositionRecordField {
	record, ok := graph.recordType(name)
	if !ok {
		return nil
	}
	var values map[string]json.RawMessage
	_ = json.Unmarshal(raw, &values)
	result := make([]CompositionRecordField, len(record.Fields))
	for i, field := range record.Fields {
		result[i] = CompositionRecordField{ID: field.ID, Name: field.Name, Value: values[field.Name]}
		if field.Presence == "optional" {
			present := values[field.Name] != nil
			result[i].Presence, result[i].Present = "optional", &present
		}
	}
	return result
}
