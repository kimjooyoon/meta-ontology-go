package workfrontier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// DecodeJSON parses one strict work-frontier input object.
func DecodeJSON(data []byte) (Input, error) {
	var input Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode work frontier JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Input{}, fmt.Errorf("decode work frontier JSON: multiple values")
		}
		return Input{}, fmt.Errorf("decode work frontier JSON: %w", err)
	}
	if input.fromJSON && input.present.minimumSelectedPressures && input.MinimumSelectedPressures < 2 {
		return Input{}, fmt.Errorf("decode work frontier JSON: minimum_selected_pressures must be at least 2")
	}
	return input, nil
}

// Decode is the JSON entry point for the v1 contract.
func Decode(data []byte) (Input, error) { return DecodeJSON(data) }

// EncodeJSON returns canonical, indented JSON terminated by one newline.
func EncodeJSON(input Input) ([]byte, error) {
	normalized := normalizeInput(input)
	encoded, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode work frontier JSON: %w", err)
	}
	return append(encoded, '\n'), nil
}

// EncodeResultJSON returns canonical JSON for a selection result.
func EncodeResultJSON(result Result) ([]byte, error) {
	normalized := normalizeResult(result)
	encoded, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode work frontier result: %w", err)
	}
	return append(encoded, '\n'), nil
}
