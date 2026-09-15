package selectiveci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode selective-ci input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode selective-ci input: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Input{}, fmt.Errorf("decode selective-ci input: trailing data")
	}
	if err := input.Validate(); err != nil {
		return Input{}, err
	}
	return normalizeInput(input), nil
}
func Decode(data []byte) (Input, error) { return DecodeJSON(data) }
func EncodeJSON(input Input) ([]byte, error) {
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	return append(canonical, '\n'), nil
}
func (input Input) CanonicalJSON() ([]byte, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
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
func EncodePlanJSON(result PlanResult) ([]byte, error) {
	result = sealResult(result)
	canonical, err := json.Marshal(normalizeResult(result))
	if err != nil {
		return nil, err
	}
	return append(canonical, '\n'), nil
}
func (result PlanResult) CanonicalJSON() ([]byte, error) {
	copy := normalizeResult(result)
	copy.CanonicalDigest = ""
	canonical, err := json.Marshal(copy)
	if err != nil {
		return nil, err
	}
	return canonical, nil
}
