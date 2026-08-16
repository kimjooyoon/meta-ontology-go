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

func (result ObligationCoverageResult) StableDigest() string {
	data, err := result.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}

func EncodeObligationCoverageJSON(result ObligationCoverageResult) ([]byte, error) {
	normalized, err := validateCoverageResult(result)
	if err != nil {
		return nil, err
	}
	digest := normalized.StableDigest()
	if normalized.OutputDigest != "" && normalized.OutputDigest != digest {
		return nil, fmt.Errorf("obligation coverage output digest mismatch")
	}
	normalized.OutputDigest = digest
	return json.Marshal(normalized)
}

func EncodeCoverageJSON(result ObligationCoverageResult) ([]byte, error) {
	return EncodeObligationCoverageJSON(result)
}

func DecodeObligationCoverageJSON(data []byte) (ObligationCoverageResult, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return ObligationCoverageResult{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var result ObligationCoverageResult
	if err := decoder.Decode(&result); err != nil {
		return ObligationCoverageResult{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ObligationCoverageResult{}, fmt.Errorf("trailing obligation coverage JSON")
	}
	normalized, err := validateCoverageResult(result)
	if err != nil {
		return ObligationCoverageResult{}, err
	}
	encoded, err := EncodeObligationCoverageJSON(normalized)
	if err != nil {
		return ObligationCoverageResult{}, err
	}
	if !bytes.Equal(encoded, data) {
		return ObligationCoverageResult{}, fmt.Errorf("non-canonical obligation coverage output")
	}
	return normalized, nil
}

func DecodeCoverageJSON(data []byte) (ObligationCoverageResult, error) {
	return DecodeObligationCoverageJSON(data)
}

func validateCoverageResult(result ObligationCoverageResult) (ObligationCoverageResult, error) {
	if result.SchemaVersion != ObligationCoverageSchemaVersion {
		return ObligationCoverageResult{}, fmt.Errorf("unsupported obligation coverage schema")
	}
	if result.Decision != CoverageDecisionExact && result.Decision != CoverageDecisionUnknown {
		return ObligationCoverageResult{}, fmt.Errorf("invalid obligation coverage decision")
	}
	if !validCoverageReason(result.Decision, result.Reason) {
		return ObligationCoverageResult{}, fmt.Errorf("invalid obligation coverage reason")
	}
	if result.FullSuiteRequired != (result.Decision == CoverageDecisionUnknown) {
		return ObligationCoverageResult{}, fmt.Errorf("inconsistent full-suite flag")
	}
	result = normalizeCoverageResult(result)
	if result.Decision == CoverageDecisionUnknown && len(result.RequiredObligationIDs) != 0 {
		return ObligationCoverageResult{}, fmt.Errorf("unknown coverage exposes required obligations")
	}
	return result, nil
}

func validCoverageReason(decision CoverageDecision, reason CoverageReason) bool {
	if decision == CoverageDecisionExact {
		return reason == CoverageReasonComplete || reason == CoverageReasonNoChange
	}
	switch reason {
	case CoverageReasonMissingInput, CoverageReasonInvalidInput, CoverageReasonUnsupportedSchema, CoverageReasonInvalidGraph,
		CoverageReasonInvalidRegistry, CoverageReasonInvalidSnapshot, CoverageReasonStaleGraph,
		CoverageReasonStaleRegistry, CoverageReasonStaleSnapshot, CoverageReasonUnknownRoot,
		CoverageReasonDuplicateRoot, CoverageReasonMissingObligation, CoverageReasonMissingCommand,
		CoverageReasonDanglingCommand, CoverageReasonWorkOverflow:
		return true
	default:
		return false
	}
}
