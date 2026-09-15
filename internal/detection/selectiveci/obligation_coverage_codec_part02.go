package selectiveci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

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
