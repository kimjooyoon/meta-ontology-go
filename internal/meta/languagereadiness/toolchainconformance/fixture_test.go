package toolchainconformance

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const fixtureHead = "1111111111111111111111111111111111111111"

func fixtureInput(t *testing.T) Input {
	t.Helper()
	registry, err := os.ReadFile("../../../../examples/toolchain-conformance/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(fixedSurfaces))
	for _, definition := range fixedSurfaces {
		indicators := make([]map[string]bool, definition.Indicators)
		for index := range indicators {
			indicators[index] = map[string]bool{"satisfied": true}
		}
		proofs := make([]map[string]bool, definition.Proofs)
		for index := range proofs {
			proofs[index] = map[string]bool{"passed": true}
		}
		value := map[string]any{
			"schema": definition.Schema, "decision": DecisionPass,
			"resolution": ResolutionExact,
			"source":     map[string]any{"expected_head_sha": fixtureHead},
			"summary": map[string]any{"total": definition.Cases,
				"satisfied": definition.Cases, "executed": definition.Cases,
				"unresolved": 0},
			"indicators": indicators, "proofs": proofs,
			"repository_writes": 0, "mutation_authorized": false,
			"report_digest": "sha256:" + strings.Repeat("a", 64),
		}
		artifacts[definition.ID], _ = json.Marshal(value)
	}
	return Input{ExpectedHeadSHA: fixtureHead, RegistryRaw: registry,
		ConceptArtifact: fixtureConcept(t), Artifacts: artifacts}
}

func fixtureConcept(t *testing.T) []byte {
	t.Helper()
	useCases := make([]map[string]string, len(expectedUseCases))
	for index, id := range expectedUseCases {
		useCases[index] = map[string]string{"id": id}
	}
	value := map[string]any{"decision": DecisionPass,
		"catalog_digest":  "sha256:" + strings.Repeat("b", 64),
		"artifact_digest": "sha256:" + strings.Repeat("c", 64),
		"report": map[string]any{"concepts": []any{map[string]any{
			"id": ExpectedConceptID, "meta_operation": ExpectedMetaOperation,
			"stage": "OPERATING", "novelty_claim": false,
			"code_bindings": expectedCodeBindings, "metric_bindings": metricIDs(),
			"use_cases": useCases,
		}}},
	}
	raw, _ := json.Marshal(value)
	return raw
}
