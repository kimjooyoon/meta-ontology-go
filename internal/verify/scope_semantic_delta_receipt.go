package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-20-semantic-delta-receipt"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/semantic-delta-receipt.yml",
		"cmd/semantic-delta-receipt-witness",
		"docs/language/semantic-delta-receipt.md",
		"examples/semantic-delta-receipt",
		"internal/meta/languageassurance/semanticdeltareceipt",
		"internal/verify/scope_semantic_delta_receipt.go",
	}
}
