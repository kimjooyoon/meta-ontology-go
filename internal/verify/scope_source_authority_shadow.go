package verify

func init() {
	branchScopeAllowlist["agent/source-authority-shadow"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/source-authority-shadow-witness",
		"examples/language-assurance-kernel/source-authority-shadow.json",
		"examples/language-assurance-kernel/source-backed-authority-shadow.md",
		"internal/meta/languageassurance/sourceauthorityshadow",
		"internal/verify/scope_source_authority_shadow.go",
	}
}
