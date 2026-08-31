package verify

func init() {
	branchScopeAllowlist["agent/source-authority-upstream"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/source-authority-upstream-witness",
		"examples/source-authority-upstream",
		"internal/meta/languageassurance/sourceauthorityupstream",
		"internal/verify/scope_source_authority_upstream.go",
	}
}
