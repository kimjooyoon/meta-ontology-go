package verify

func init() {
	branchScopeAllowlist["agent/external-assurance-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/external-conformance-activation-witness",
		"examples/external-conformance-activation",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/externalconformanceactivation",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/verify/scope_external_conformance_activation.go",
	}
}
