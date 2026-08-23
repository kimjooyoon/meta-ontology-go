package verify

func init() {
	branchScopeAllowlist["agent/language-assurance-kernel"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-assurance.yml",
		"cmd/language-assurance-witness",
		"docs/language/language-assurance-kernel.md",
		"examples/language-assurance-kernel",
		"internal/meta/languageassurance",
		"internal/verify/scope_language_assurance.go",
	}
}
