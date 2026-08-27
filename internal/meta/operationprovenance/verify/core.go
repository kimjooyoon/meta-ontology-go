package verify

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const receiptSchema = "gooo/meta-operation-provenance-receipt/v2"
const toolchain = "go1.27.0"

// Verify is the independent consumer entry point. It starts at raw source.
func Verify(payload, source, consumerSource []byte) (map[string]any, error) {
	imports := producerImportCheck(consumerSource)
	if len(source) == 0 {
		return unknownReport(imports), nil
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
	metrics, scenarios, reconstruction, err := reconstruct(ir)
	if err != nil {
		return nil, err
	}
	expected := receipt{Schema: actual.Schema, Toolchain: actual.Toolchain, Source: actual.Source, Semantic: actual.Semantic, Reconstruction: reconstruction, Observation: actual.Observation}
	for _, scenario := range scenarios {
		expected.Scenarios = append(expected.Scenarios, evaluate(scenario, metrics, actual.Source, actual.Semantic))
	}
	if !reflect.DeepEqual(actual.Scenarios, expected.Scenarios) || actual.Reconstruction != expected.Reconstruction {
		return nil, fmt.Errorf("receipt differs from independent semantic reconstruction")
	}
	withoutDigest := actual
	withoutDigest.Digest = ""
	wantDigest, err := digestJSON(withoutDigest)
	if err != nil || actual.Digest != wantDigest {
		return nil, fmt.Errorf("receipt digest is not bound")
	}
	return verifiedReport(actual, metrics, imports)
}

func verifyHeader(actual receipt, source []byte, ir semantic.IR) error {
	if actual.Schema != receiptSchema || actual.Toolchain != toolchain {
		return fmt.Errorf("receipt schema or toolchain is invalid")
	}
	if actual.Source != digest(source) || actual.Semantic != "sha256:"+ir.StableHash() {
		return fmt.Errorf("raw or canonical semantic source digest is not bound")
	}
	if actual.Observation.Before != actual.Observation.After || len(actual.Observation.Changed) != 0 || actual.Observation.Writes || actual.Observation.Authority {
		return fmt.Errorf("repository write observation is not clean")
	}
	return nil
}
