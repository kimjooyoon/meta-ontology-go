package valueexecution

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

// DecodeRecordInput rejects ambiguous JSON rather than choosing the last
// duplicate field. It decodes scalars and ordered scalar arrays; the schema determines
// which fields may be omitted. JSON null is never an absent value.
func DecodeRecordInput(raw []byte) (RecordFields, error) {
	failure := failAt(ReasonExternalInputUnexpected, "INPUT", "decode-record-input", "input must be one object with unique scalar or ordered scalar-array fields")
	if !utf8.Valid(raw) || !json.Valid(raw) {
		return nil, failure
	}
	if err := valuejson.CheckEscapesInValidJSON(raw); err != nil {
		return nil, failure
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, failure
	}
	fields := RecordFields{}
	for decoder.More() {
		key, err := decoder.Token()
		name, isString := key.(string)
		if err != nil || !isString {
			return nil, failure
		}
		if _, duplicate := fields[name]; duplicate {
			return nil, failure
		}
		var value any
		if err := decoder.Decode(&value); err != nil {
			return nil, failure
		}
		value, err = decodeRecordFieldValue(value)
		if err != nil {
			return nil, failure
		}
		fields[name] = value
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, failure
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, failure
	}
	return fields, nil
}
