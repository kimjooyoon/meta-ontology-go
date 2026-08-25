package verify

func init() {
	branchScopeAllowlist["agent/gooo-observation-input"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-observation.yml",
		"examples/self-improvement",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/selfimprovementobservation",
		"internal/verify/scope_gooo_observation_input.go",
		"scripts/self-improvement-contract",
		"scripts/self-improvement-cycle",
		"scripts/self-improvement-observation",
	}
}
