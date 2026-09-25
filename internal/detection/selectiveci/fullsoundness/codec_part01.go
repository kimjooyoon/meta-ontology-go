package fullsoundness

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode full soundness input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode full soundness input: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Input{}, fmt.Errorf("decode full soundness input: %w", err)
	}
	return normalizeInput(input), nil
}
func ClassifyJSON(data []byte) Output {
	input, err := DecodeJSON(data)
	if err != nil {
		return seal(Output{SchemaVersion: SchemaVersion}, DecisionUnknown, ReasonFullSuiteRequired)
	}
	return Evaluate(input)
}
func EncodeInputJSON(input Input) ([]byte, error) {
	data, err := json.Marshal(normalizeInput(input))
	if err != nil {
		return nil, fmt.Errorf("encode full soundness input: %w", err)
	}
	return append(data, '\n'), nil
}
func (input Input) CanonicalJSON() ([]byte, error) { return json.Marshal(normalizeInput(input)) }
func EncodeJSON(output Output) ([]byte, error) {
	output = normalizeOutput(output)
	decisionDigest, err := output.DecisionStableDigest()
	if err != nil {
		return nil, fmt.Errorf("encode full soundness decision: %w", err)
	}
	if output.DecisionDigest != "" && output.DecisionDigest != decisionDigest {
		return nil, fmt.Errorf("full soundness decision digest mismatch")
	}
	output.DecisionDigest = decisionDigest
	envelopeDigest := output.StableDigest()
	if output.CanonicalDigest != "" && output.CanonicalDigest != envelopeDigest {
		return nil, fmt.Errorf("full soundness envelope digest mismatch")
	}
	output.CanonicalDigest = envelopeDigest
	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("encode full soundness output: %w", err)
	}
	return append(data, '\n'), nil
}
func (output Output) CanonicalJSON() ([]byte, error) {
	output = normalizeOutput(output)
	output.CanonicalDigest = ""
	return json.Marshal(output)
}
func (output Output) StableDigest() string {
	data, err := output.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}
