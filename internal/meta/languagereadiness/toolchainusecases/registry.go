package toolchainusecases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

func decodeRegistry(raw []byte) (Registry, error) {
	registry := Registry{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return registry, fmt.Errorf("decode executable use case registry: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return registry, fmt.Errorf("decode executable use case registry: trailing content")
	}
	if !reflect.DeepEqual(registry, expectedRegistry()) {
		return registry, fmt.Errorf("executable use case registry mismatch")
	}
	return registry, nil
}

func unresolvedCases(artifactDigest string) []CaseResult {
	definitions := expectedRegistry().Cases
	results := make([]CaseResult, 0, len(definitions))
	for _, definition := range definitions {
		item := CaseResult{Definition: definition, ObservedDecision: "UNKNOWN", Status: "UNRESOLVED"}
		item.EvidenceDigest = caseDigest(item, artifactDigest)
		results = append(results, item)
	}
	return results
}
