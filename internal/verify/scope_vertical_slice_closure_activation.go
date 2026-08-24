package verify

func init() {
	branchScopeAllowlist["agent/vertical-slice-closure-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/vertical-slice-closure-activation-witness",
		"examples/vertical-slice-closure-activation",
		"internal/meta/languageassurance/digest.go",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/meta/languageassurance/verticalsliceclosureactivation",
		"internal/verify/scope_vertical_slice_closure_activation.go",
	}
}
