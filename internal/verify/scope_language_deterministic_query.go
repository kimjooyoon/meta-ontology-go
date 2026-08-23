package verify

func init() {
	branchScopeAllowlist["agent/language-deterministic-query"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"bootstrap/function-extractor/recipes.json",
		"cmd/language-deterministic-query-readiness-binding",
		"cmd/language-deterministic-query-witness",
		"cmd/language-readiness-witness/run_test.go",
		"docs/language/language-deterministic-query-readiness-binding.md",
		"docs/language/language-deterministic-query.md",
		"examples/language-deterministic-query",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/languagedeterministicquery",
		"internal/meta/languagereadiness/languagedeterministicquerybinding",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/verify/scope_language_deterministic_query.go",
	}
}
