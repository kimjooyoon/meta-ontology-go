package bindingcoverage

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func ClassifyJSON(data []byte) Output {
	input, err := DecodeJSON(data)
	if err != nil {
		output := Output{SchemaVersion: SchemaVersion, InputBytes: uint64(len(data))}
		if schema, ok := schemaFromJSON(data); ok && schema != SchemaVersion {
			output.SchemaVersion = schema
			return seal(output, DecisionUnknown, ReasonUnknownSchema)
		}
		return seal(output, DecisionUnknown, ReasonMissingInput)
	}
	return Observe(input)
}
func ObserveJSON(data []byte) Output { return ClassifyJSON(data) }
func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode binding coverage input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode binding coverage input: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Input{}, fmt.Errorf("decode binding coverage input: %w", err)
	}
	return normalizeInput(input), nil
}
func Decode(data []byte) (Input, error) { return DecodeJSON(data) }
func EncodeInputJSON(input Input) ([]byte, error) {
	data, err := json.Marshal(normalizeInput(input))
	if err != nil {
		return nil, fmt.Errorf("encode binding coverage input: %w", err)
	}
	return append(data, '\n'), nil
}
func (input Input) CanonicalJSON() ([]byte, error) {
	return json.Marshal(normalizeInput(input))
}
func (input Input) Canonical() string {
	data, err := input.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(data)
}
func (input Input) Digest() string { return digestBytes([]byte(input.Canonical())) }
func EncodeJSON(output Output) ([]byte, error) {
	output = normalizeOutput(output)
	output.CanonicalDigest = output.StableDigest()
	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("encode binding coverage output: %w", err)
	}
	return append(data, '\n'), nil
}
func (output Output) CanonicalJSON() ([]byte, error) {
	canonical := normalizeOutput(output)
	canonical.CanonicalDigest = ""
	return json.Marshal(canonical)
}
