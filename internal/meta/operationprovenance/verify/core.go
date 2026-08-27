package verify

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
)

// Verify independently parses source and reads every raw lineage artifact.
func Verify(payload, source, consumerSource []byte, artifactRoots ...string) (map[string]any, error) {
	imports := producerImportCheck(consumerSource)
	if len(source) == 0 {
		return unknownReport(imports), nil
	}
	root := filepath.Join("examples", "meta-operation-provenance", "artifacts")
	if len(artifactRoots) > 0 && artifactRoots[0] != "" {
		root = artifactRoots[0]
	}
	var actual receipt
	if err := json.Unmarshal(payload, &actual); err != nil {
		return nil, fmt.Errorf("decode receipt: %w", err)
	}
	ir, err := lower(source)
	if err != nil {
		return nil, err
	}
	if err := verifyHeader(actual, source, ir); err != nil {
		return nil, err
	}
	metrics, scenarios, reconstruction, sourceIssues, err := reconstruct(ir)
	if err != nil {
		return nil, err
	}
	families, err := validateContract(metrics, scenarios)
	if err != nil {
		return nil, err
	}
	artifacts := collectArtifacts(root, metrics)
	expected := receipt{Schema: actual.Schema, Toolchain: actual.Toolchain, Source: actual.Source, Semantic: actual.Semantic, SourceResolution: actual.SourceResolution, Reconstruction: reconstruction, SourceIssues: sourceIssues, FamilyCardinality: families, Observation: actual.Observation}
	for _, scenario := range scenarios {
		result, err := evaluate(scenario, metrics, artifacts, actual.Source, actual.Semantic)
		if err != nil {
			return nil, err
		}
		result.SourceResolution = actual.SourceResolution
		expected.Scenarios = append(expected.Scenarios, result)
	}
	if !reflect.DeepEqual(actual.Scenarios, expected.Scenarios) || actual.Reconstruction != expected.Reconstruction || !reflect.DeepEqual(actual.SourceIssues, expected.SourceIssues) || !reflect.DeepEqual(actual.FamilyCardinality, expected.FamilyCardinality) {
		return nil, fmt.Errorf("receipt differs from independent semantic and artifact reconstruction")
	}
	withoutDigest := actual
	withoutDigest.Digest = ""
	wantDigest, err := digestJSON(withoutDigest)
	if err != nil || actual.Digest != wantDigest {
		return nil, fmt.Errorf("receipt digest is not bound")
	}
	return verifiedReport(actual, metrics, imports), nil
}
