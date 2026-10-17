package valuejson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

// Object retains raw field values and refuses repeated names before decoding
// them into a map. The caller supplies its existing field-count budget.
func Object(raw []byte, maximum int) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) || !json.Valid(raw) {
		return nil, fmt.Errorf("explicit JSON object required")
	}
	if err := CheckEscapesInValidJSON(raw); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, fmt.Errorf("explicit JSON object required")
	}
	values := make(map[string]json.RawMessage)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := key.(string)
		if !ok || values[name] != nil || len(values) >= maximum {
			return nil, fmt.Errorf("record fields must be distinct names, at most %d", maximum)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		values[name] = value
	}
	if end, err := decoder.Token(); err != nil || end != json.Delim('}') {
		return nil, fmt.Errorf("record object is incomplete")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("record object must contain one value")
	}
	return values, nil
}
