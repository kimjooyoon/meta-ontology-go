package verify

func init() {
	branchScopeAllowlist["agent/gooo-ci-plan"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci-plan-usecase.yml",
		"cmd/ci-plan-scorecard",
		"cmd/gooo",
		"examples/ci-plan",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/ciplanusecase",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/metainvocation",
		"internal/verify/scope_ci_plan_usecase.go",
		"scripts/ci-plan-usecase",
	}
}
