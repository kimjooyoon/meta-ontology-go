package verify

func init() {
	paths := []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/symbolic-invocation-schema",
		"examples/symbolic-invocation-usecase",
		"internal/meta/symbolicinvocationusecase",
		"internal/packageruntime/artifactemit",
		"internal/verify/scope_gooo_closed_generation_example.go",
		"scripts/symbolic-invocation-schema",
		"scripts/symbolic-invocation-usecase",
	}
	branchScopeAllowlist["agent/gooo-closed-generation-example"] = paths
	branchScopeAllowlist["agent/gooo-reader-resolution-projection"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"examples/language-syntax-roundtrip/README.md",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/symbolic-invocation-schema",
		"examples/symbolic-invocation-usecase",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/symbolicinvocationusecase",
		"internal/packageruntime/artifactemit",
		"internal/verify/scope_gooo_closed_generation_example.go",
		"scripts/symbolic-invocation-schema",
		"scripts/symbolic-invocation-usecase",
	}
}
