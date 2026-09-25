package verify

func init() {
	branchScopeAllowlist["agent/changed-surface-receipt-eligibility"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/changed-surface-receipt-eligibility-witness",
		"examples/changed-surface-receipt-eligibility",
		"internal/meta/languageassurance/changedsurfacereceipteligibility",
		"internal/verify/scope_changed_surface_receipt_eligibility.go",
	}
}
