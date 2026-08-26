package artifactemit

import (
	"encoding/json"
	"testing"
)

func symbolicValueReachabilityFixture(t *testing.T) []byte {
	t.Helper()
	encoded, err := json.Marshal(symbolicValueReachabilityFixtureMap())
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func symbolicValueReachabilityFixtureMap() map[string]any {
	fixture := symbolicValueContractFixtureMap()
	fixture["resolution"] = "SYMBOLIC_ONLY"
	fixture["kind"] = "symbolic-invocation-schema"
	fixture["effects"] = map[string]any{"repository_writes": 0, "mutation_authority": false}
	fixture["json_schema"] = map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title":   "Checkout symbolic invocation",
		"type":    "object",
		"properties": map[string]any{
			"activity": map[string]any{"const": "Checkout"},
			"inputs": map[string]any{
				"type": "array", "prefixItems": []any{
					map[string]any{"const": "urn:gooo:checkout:cart"},
					map[string]any{"const": "urn:gooo:checkout:payment-method"},
				}, "items": false, "minItems": 2, "maxItems": 2,
			},
		},
		"examples": []any{map[string]any{
			"activity": "Checkout",
			"inputs": []any{"urn:gooo:checkout:cart", "urn:gooo:checkout:payment-method"},
		}},
		"required": []any{"activity", "inputs"}, "additionalProperties": false,
	}
	return fixture
}

func symbolicValueReachabilityContract(t *testing.T, artifact []byte, subjectSHA string) []byte {
	t.Helper()
	contract, err := CompileSymbolicValueContract(artifact, subjectSHA)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
