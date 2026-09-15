package impactcoverage

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func ObserveJSON(data []byte) Result {
	input, err := DecodeInput(data)
	if err != nil {
		result := Result{Schema: SchemaV1, ChangedStableIDs: []string{}, UncoveredPaths: []string{}}
		return seal(result, DecisionUnknown, ReasonInvalidSnapshot)
	}
	return Observe(input)
}
func EvaluateJSON(data []byte) Result             { return ObserveJSON(data) }
func Decode(data []byte) (Input, error)           { return DecodeInput(data) }
func EncodeInputJSON(input Input) ([]byte, error) { return input.CanonicalJSON() }
func (result Result) CanonicalJSON() ([]byte, error) {
	normalized, err := normalizeResult(result)
	if err != nil {
		return nil, err
	}
	normalized.OutputDigest = ""
	return json.Marshal(normalized)
}
func (result Result) StableDigest() string {
	data, err := result.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}
func EncodeJSON(result Result) ([]byte, error) {
	normalized, err := normalizeResult(result)
	if err != nil {
		return nil, err
	}
	digest := normalized.StableDigest()
	if normalized.OutputDigest != "" && normalized.OutputDigest != digest {
		return nil, fmt.Errorf("impact coverage output digest mismatch")
	}
	normalized.OutputDigest = digest
	return json.Marshal(normalized)
}
func DecodeJSON(data []byte) (Result, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Result{}, fmt.Errorf("decode impact coverage output: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var result Result
	if err := decoder.Decode(&result); err != nil {
		return Result{}, fmt.Errorf("decode impact coverage output: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Result{}, fmt.Errorf("decode impact coverage output: %w", err)
	}
	normalized, err := normalizeResult(result)
	if err != nil {
		return Result{}, err
	}
	encoded, err := EncodeJSON(normalized)
	if err != nil {
		return Result{}, err
	}
	if !bytes.Equal(encoded, data) {
		return Result{}, fmt.Errorf("impact coverage output is not canonical")
	}
	return normalized, nil
}
