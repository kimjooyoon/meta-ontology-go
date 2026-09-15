package selectiveci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func normalizeCoverageInput(input ObligationCoverageInput) ObligationCoverageInput {
	if graph, err := input.Graph.Normalized(); err == nil {
		input.Graph = graph
	}
	input.Registry = normalizeRegistry(input.Registry)
	input.ChangedRootIDs = sortedCopy(input.ChangedRootIDs)
	return input
}
func (input ObligationCoverageInput) CanonicalJSON() ([]byte, error) {
	if input.SchemaVersion != ObligationCoverageSchemaVersion {
		return nil, fmt.Errorf("unsupported obligation coverage schema")
	}
	return json.Marshal(normalizeCoverageInput(input))
}
func (input ObligationCoverageInput) Digest() string {
	data, err := input.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}
func EncodeObligationCoverageInputJSON(input ObligationCoverageInput) ([]byte, error) {
	data, err := input.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
func DecodeObligationCoverageInputJSON(data []byte) (ObligationCoverageInput, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return ObligationCoverageInput{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input ObligationCoverageInput
	if err := decoder.Decode(&input); err != nil {
		return ObligationCoverageInput{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ObligationCoverageInput{}, fmt.Errorf("trailing obligation coverage input JSON")
	}
	return normalizeCoverageInput(input), nil
}
func ObserveObligationCoverageJSON(data []byte) ObligationCoverageResult {
	input, err := DecodeObligationCoverageInputJSON(data)
	if err != nil {
		return sealCoverage(ObligationCoverageResult{SchemaVersion: ObligationCoverageSchemaVersion}, CoverageDecisionUnknown, CoverageReasonInvalidInput)
	}
	return ObserveObligationCoverage(input)
}
func EvaluateObligationCoverageJSON(data []byte) ObligationCoverageResult {
	return ObserveObligationCoverageJSON(data)
}
func (result ObligationCoverageResult) CanonicalJSON() ([]byte, error) {
	normalized, err := validateCoverageResult(result)
	if err != nil {
		return nil, err
	}
	normalized.OutputDigest = ""
	return json.Marshal(normalized)
}
