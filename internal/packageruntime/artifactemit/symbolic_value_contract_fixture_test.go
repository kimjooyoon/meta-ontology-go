package artifactemit

import (
	"encoding/json"
	"strings"
	"testing"
)

func symbolicValueContractFixture(t *testing.T) []byte {
	t.Helper()
	artifact, err := json.Marshal(symbolicValueContractFixtureMap())
	if err != nil {
		t.Fatal(err)
	}
	return artifact
}

func symbolicValueContractFixtureMap() map[string]any {
	inputs := []any{"urn:gooo:checkout:cart", "urn:gooo:checkout:payment-method"}
	return map[string]any{
		"schema": "gooo/symbolic-invocation-schema-artifact/v1", "decision": "PASS",
		"digest": "sha256:" + strings.Repeat("a", 64),
		"conformance": map[string]any{
			"schema": "gooo/symbolic-invocation-conformance/v1", "decision": "PASS",
			"resolution": "STRUCTURAL_ONLY", "generated_vectors": 2,
			"embedded_handwritten_vectors": 0,
			"vectors": []any{
				symbolicValueFixtureVector("accept-exact", "ACCEPT", "FOUNDATION", "project-exact-symbolic-invocation", "Checkout", inputs),
				symbolicValueFixtureVector("reject-missing-activity", "REJECT", "REGRESSION", "remove-required-activity", "", inputs),
			},
			"effects": map[string]any{"repository_writes": 0, "mutation_authority": false},
		},
	}
}

func symbolicValueFixtureVector(id, expected, proof, operation, activity string, inputs []any) map[string]any {
	instance := map[string]any{"inputs": inputs}
	if activity != "" {
		instance["activity"] = activity
	}
	return map[string]any{
		"id": id, "expected": expected, "proof_choice": proof,
		"meta_operation": operation, "instance": instance,
	}
}

func fixtureVectors(fixture map[string]any) []map[string]any {
	raw := fixture["conformance"].(map[string]any)["vectors"].([]any)
	return []map[string]any{raw[0].(map[string]any), raw[1].(map[string]any)}
}
