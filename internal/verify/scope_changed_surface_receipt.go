package verify

func init() {
	branchScopeAllowlist["agent/changed-surface-receipt-shadow"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/changed-surface-receipt-witness",
		"examples/changed-surface-receipt-shadow",
		"internal/meta/languageassurance/changedsurfacereceipt",
		"internal/verify/scope_changed_surface_receipt.go",
	}
}
