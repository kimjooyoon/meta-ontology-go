package valueexecution

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func decodeRecordFieldValue(value any) (any, error) {
	if values, ok := value.([]any); ok {
		if values == nil || len(values) > syntax.RepeatedFieldLimit {
			return nil, fmt.Errorf("repeated field exceeds its ordered value profile")
		}
		for i, item := range values {
			decoded, err := decodeRecordScalar(item)
			if err != nil {
				return nil, err
			}
			values[i] = decoded
		}
		return values, nil
	}
	return decodeRecordScalar(value)
}

func decodeRecordScalar(value any) (any, error) {
	switch typed := value.(type) {
	case string, bool:
		return value, nil
	case json.Number:
		return typed.Int64()
	default:
		return nil, fmt.Errorf("record element requires Text, Boolean or Integer")
	}
}

func recordFieldValueMatches(field RecordField, value any) bool {
	if field.TextBound.Validate() != nil || field.TextBound.IsSet() && field.TypeID != string(semantic.BuiltinStringTypeID) {
		return false
	}
	if field.Cardinality == semantic.Many {
		values, ok := value.([]any)
		if !ok || values == nil || len(values) > syntax.RepeatedFieldLimit {
			return false
		}
		for _, item := range values {
			if !recordBoundedScalarMatches(field, item) {
				return false
			}
		}
		return true
	}
	return recordBoundedScalarMatches(field, value)
}

func recordBoundedScalarMatches(field RecordField, value any) bool {
	if !recordScalarMatches(field.TypeID, value) {
		return false
	}
	if !field.TextBound.IsSet() {
		return true
	}
	maximum, err := field.TextBound.Resolve(0)
	text, ok := value.(string)
	return err == nil && ok && uint64(len(text)) <= uint64(maximum)
}

func recordScalarMatches(typeID string, value any) bool {
	switch typeID {
	case string(semantic.BuiltinStringTypeID):
		text, ok := value.(string)
		return ok && utf8.ValidString(text)
	case string(semantic.BuiltinBooleanTypeID):
		_, ok := value.(bool)
		return ok
	case string(semantic.BuiltinIntegerTypeID):
		_, ok := value.(int64)
		return ok
	default:
		return false
	}
}

func cloneRecordFields(fields RecordFields) RecordFields {
	copy := maps.Clone(fields)
	for name, value := range copy {
		if values, ok := value.([]any); ok {
			copy[name] = slices.Clone(values)
		}
	}
	return copy
}
