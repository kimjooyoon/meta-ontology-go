package verify

func init() {
	branchScopeAllowlist["agent/source-authority-promotion"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/source-authority-promotion-witness",
		"examples/source-authority-promotion",
		"internal/meta/languageassurance/sourceauthoritypromotion",
		"internal/verify/scope_source_authority_promotion.go",
	}
}
