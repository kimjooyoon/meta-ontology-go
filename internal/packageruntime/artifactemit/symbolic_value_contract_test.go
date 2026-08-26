package artifactemit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompileSymbolicValueContract(t *testing.T) {
	artifact := symbolicValueContractFixture(t)
	contract, err := CompileSymbolicValueContract(artifact, strings.Repeat("b", 40))
	if err != nil {
		t.Fatalf("CompileSymbolicValueContract() error = %v", err)
	}
	if contract.Decision != "PASS" || contract.Resolution != "VALUE_CONTRACT_ONLY" {
		t.Fatalf("contract decision = %s/%s", contract.Decision, contract.Resolution)
	}
	if contract.Coordinates.Satisfied != 8 || contract.Coordinates.Total != 8 {
		t.Fatalf("contract coordinates = %d/%d", contract.Coordinates.Satisfied, contract.Coordinates.Total)
	}
	if len(contract.Rules) != 2 || contract.Rules[0].Decision != "READY" || contract.Rules[1].Decision != "FAIL_CLOSED" {
		t.Fatalf("contract rules = %#v", contract.Rules)
	}
	if contract.Default.Decision != "FAIL_CLOSED" || !strings.HasPrefix(contract.Digest, "sha256:") {
		t.Fatalf("contract default/digest = %#v %q", contract.Default, contract.Digest)
	}
	replay, err := CompileSymbolicValueContract(artifact, strings.Repeat("b", 40))
	if err != nil {
		t.Fatalf("CompileSymbolicValueContract() replay error = %v", err)
	}
	if replay.Digest != contract.Digest {
		t.Fatalf("contract digest replay = %q, want %q", replay.Digest, contract.Digest)
	}
}

func TestCompileSymbolicValueContractRejectsUnknownExpectedVerdict(t *testing.T) {
	fixture := symbolicValueContractFixtureMap(t)
	fixture["conformance"].(map[string]any)["vectors"].([]any)[0].(map[string]any)["expected"] = "UNKNOWN"
	artifact, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompileSymbolicValueContract(artifact, strings.Repeat("b", 40)); err == nil {
		t.Fatal("CompileSymbolicValueContract() accepted an unknown expected verdict")
	}
}

func TestCompileSymbolicValueContractRejectsIncompleteAcceptVector(t *testing.T) {
	fixture := symbolicValueContractFixtureMap(t)
	fixture["conformance"].(map[string]any)["vectors"].([]any)[0].(map[string]any)["instance"].(map[string]any)["activity"] = ""
	artifact, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompileSymbolicValueContract(artifact, strings.Repeat("b", 40)); err == nil {
		t.Fatal("CompileSymbolicValueContract() accepted an incomplete accept vector")
	}
}

func symbolicValueContractFixture(t *testing.T) []byte {
	t.Helper()
	artifact, err := json.Marshal(symbolicValueContractFixtureMap(t))
	if err != nil {
		t.Fatal(err)
	}
	return artifact
}

func symbolicValueContractFixtureMap(t *testing.T) map[string]any {
	t.Helper()
	inputs := []any{"urn:gooo:checkout:cart", "urn:gooo:checkout:payment-method"}
	return map[string]any{
		"schema":   "gooo/symbolic-invocation-schema-artifact/v1",
		"decision": "PASS",
		"digest":   "sha256:" + strings.Repeat("a", 64),
		"conformance": map[string]any{
			"schema":                       "gooo/symbolic-invocation-conformance/v1",
			"decision":                     "PASS",
			"resolution":                   "STRUCTURAL_ONLY",
			"generated_vectors":            2,
			"embedded_handwritten_vectors": 0,
			"vectors": []any{
				map[string]any{
					"id":             "accept-exact",
					"expected":       "ACCEPT",
					"proof_choice":   "FOUNDATION",
					"meta_operation": "project-exact-symbolic-invocation",
					"instance": map[string]any{
						"activity": "Checkout",
						"inputs":   inputs,
					},
				},
				map[string]any{
					"id":             "reject-missing-activity",
					"expected":       "REJECT",
					"proof_choice":   "REGRESSION",
					"meta_operation": "remove-required-activity",
					"instance": map[string]any{
						"inputs": inputs,
					},
				},
			},
			"effects": map[string]any{
				"repository_writes": 0,
				"mutation_authority": false,
			},
		},
	}
}
