package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-16-meta-circular-boundary"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/meta-circular-boundary.yml",
		".github/workflows/self-improvement-contract.yml",
		"cmd/meta-circular-boundary-consumer-witness",
		"cmd/meta-circular-boundary-intervention-witness",
		"cmd/meta-circular-boundary-witness",
		"examples/language-semantic-model/corpus.json",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/meta-circular-boundary",
		"examples/toolchain-conformance/corpus.json",
		"internal/meta/languagereadiness/languagesemantic/model.go",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/languagereadiness/toolchainconformance/contract.go",
		"internal/meta/languagereadiness/toolchainconformance/corpus.go",
		"internal/meta/metacircularboundary",
		"internal/meta/metacircularboundarycontract",
		"internal/meta/metacircularboundaryconsumer",
		"internal/verify/scope_meta_circular_boundary.go",
	}
}
