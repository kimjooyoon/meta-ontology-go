package verify

func init() {
	branchScopeAllowlist["agent/gooo-operation-catalog-kernel"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-operation-catalog.yml",
		"examples/language-operation-catalog",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/valuecatalog",
		"internal/verify/scope_gooo_operation_catalog.go",
		"scripts/language-operation-catalog",
	}
	branchScopeAllowlist["agent/gooo-operation-catalog-extension"] = []string{
		"examples/language-operation-catalog/main.gooo",
	}
	branchScopeAllowlist["agent/gooo-operation-spec-ir"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-operation-catalog.yml",
		"examples/language-operation-catalog",
		"internal/valuecatalog",
		"internal/valueexecution",
		"internal/verify/scope_gooo_operation_catalog.go",
		"scripts/language-operation-catalog",
	}
	branchScopeAllowlist["agent/gooo-operation-claim-transitions"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-operation-catalog.yml",
		"examples/language-operation-catalog",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/valuecatalog",
		"internal/verify/scope_gooo_operation_catalog.go",
		"scripts/language-operation-catalog",
	}
}
