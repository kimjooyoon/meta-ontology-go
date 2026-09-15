package lanefrontier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// ClassifyJSON strictly decodes one input value. Malformed or incomplete
// input becomes the same fail-closed UNKNOWN result as missing facts.
func ClassifyJSON(data []byte) Output {
	input, err := DecodeJSON(data)
	if err != nil {
		if schema, ok := schemaFromJSON(data); ok && schema != SchemaVersion {
			return seal(Output{SchemaVersion: schema}, DecisionUnknown, ReasonUnknownSchema)
		}
		return seal(Output{SchemaVersion: SchemaVersion}, DecisionUnknown, ReasonMissingInput)
	}
	return Classify(input)
}

// DecodeJSON accepts exactly one versioned input object with no duplicate or
// unknown fields. Semantic eligibility is evaluated separately by Classify.
func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode lane frontier input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode lane frontier input: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Input{}, fmt.Errorf("decode lane frontier input: trailing data")
	}
	return normalizeInput(input), nil
}
func Decode(data []byte) (Input, error) { return DecodeJSON(data) }

// EncodeInputJSON emits deterministic input JSON. Classify remains the
// authority for deciding whether the facts are complete and eligible.
func EncodeInputJSON(input Input) ([]byte, error) {
	data, err := json.Marshal(normalizeInput(input))
	if err != nil {
		return nil, fmt.Errorf("encode lane frontier input: %w", err)
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
