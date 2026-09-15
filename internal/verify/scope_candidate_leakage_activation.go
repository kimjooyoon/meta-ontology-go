package verify

func init() {
	branchScopeAllowlist["agent/candidate-leakage-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/candidate-leakage-activation-witness",
		"examples/candidate-leakage-activation",
		"internal/meta/languageassurance/candidateleakageactivation",
		"internal/meta/languageassurance/digest.go",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/verify/scope_candidate_leakage_activation.go",
	}
}
