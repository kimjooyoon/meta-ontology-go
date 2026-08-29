package verify

const opentofuObservationBranch = "agent/opentofu-observation-v1"

func init() {
	branchScopeAllowlist[opentofuObservationBranch] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/opentofu-observation.yml",
		"cmd/opentofu-observation-witness",
		"docs/external/opentofu-observation-v1.md",
		"examples/opentofu-observation",
		"examples/language-semantic-model/corpus.json",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/opentofuobservation",
		"internal/meta/languagereadiness/languagesemantic/model.go",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/verify/scope_opentofu_observation_v1.go",
		"internal/verify/scope_opentofu_observation_v1_test.go",
		"scripts/opentofu-observation",
	}
}
