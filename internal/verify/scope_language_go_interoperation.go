package verify

func init() {
	branchScopeAllowlist["agent/language-go-interoperation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-go-interoperation-readiness-binding",
		"cmd/language-go-interoperation-witness",
		"cmd/language-readiness-witness/run_test.go",
		"docs/language/language-go-interoperation-readiness-binding.md",
		"docs/language/language-go-interoperation.md",
		"examples/language-go-interoperation",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/languagegointeroperation",
		"internal/meta/languagereadiness/languagegointeroperationbinding",
		"internal/verify/scope_language_go_interoperation.go",
	}
}
