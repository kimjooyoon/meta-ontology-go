package verify

func init() {
	branchScopeAllowlist["agent/gooo-source-binding-promotion"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-source-binding-promotion",
		"examples/language-source-binding-promotion",
		"examples/self-improvement/main.gooo",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagesourcebindingpromotion",
		"internal/meta/languageconcept",
		"internal/verify/scope_language_source_binding_promotion.go",
		"scripts/language-source-binding-promotion",
	}
}
