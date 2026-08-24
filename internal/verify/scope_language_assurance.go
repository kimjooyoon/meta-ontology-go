package verify

func init() {
	branchScopeAllowlist["agent/language-assurance-kernel"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"bootstrap/function-extractor/recipes.json",
		"cmd/language-assurance-witness",
		"docs/language/language-assurance-kernel.md",
		"examples/language-assurance-kernel",
		"internal/meta/languageassurance",
		"internal/verify/scope_language_assurance.go",
	}
}
