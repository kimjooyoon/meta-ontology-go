package verify

func init() {
	branchScopeAllowlist["agent/language-semantic-model"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness/run_test.go",
		"cmd/language-semantic-readiness-binding",
		"cmd/language-semantic-witness",
		"docs/language/language-semantic-model.md",
		"docs/language/language-semantic-readiness-binding.md",
		"examples/language-semantic-model",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/verify/scope_language_semantic_model.go",
	}
}
