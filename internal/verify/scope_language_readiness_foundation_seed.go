package verify

func init() {
	branchScopeAllowlist["agent/language-readiness-foundation-seed"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness/foundation-seed",
		"examples/language-semantic-model/corpus.json",
		"internal/meta/languagereadiness/artifact/foundationseed",
		"internal/meta/languagereadiness/languagesemantic/evaluate_test.go",
		"internal/meta/languagereadiness/languagesemantic/model.go",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/verify/scope_language_readiness_foundation_seed.go",
	}
}
