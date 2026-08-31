package verify

func init() {
	branchScopeAllowlist["agent/gooo-source-binding-promotion"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-source-binding-promotion",
		"examples/language-source-binding-promotion",
		"examples/self-improvement/main.gooo",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagesourcebindingpromotion",
		"internal/meta/languageconcept",
		"internal/verify/scope_language_source_binding_promotion.go",
		"scripts/language-source-binding-promotion",
		"scripts/self-improvement-contract/contract_test.go",
	}
}
