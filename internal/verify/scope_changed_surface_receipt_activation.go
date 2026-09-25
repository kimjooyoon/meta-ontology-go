package verify

func init() {
	branchScopeAllowlist["agent/changed-surface-receipt-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/changed-surface-receipt-activation-witness",
		"examples/changed-surface-receipt-activation",
		"internal/meta/languageassurance/changedsurfacereceiptactivation",
		"internal/meta/languageassurance/digest.go",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/verify/scope_changed_surface_receipt_activation.go",
	}
}
