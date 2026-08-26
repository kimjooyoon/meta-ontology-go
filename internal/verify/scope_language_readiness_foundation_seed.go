package verify

func init() {
	branchScopeAllowlist["agent/language-readiness-foundation-seed"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness/foundation-seed",
		"examples/language-semantic-model/corpus.json",
		"examples/toolchain-conformance/corpus.json",
		"examples/vertical-slice-closure-shadow/README.md",
		"internal/meta/languageassurance/verticalsliceclosureshadow/artifact_link_values.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/artifact_values.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/languageassurance/verticalsliceclosureshadow/fixture_language_test.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/fixture_toolchain_test.go",
		"internal/meta/languagereadiness/artifact/foundationseed",
		"internal/meta/languagereadiness/languagesemantic/evaluate_test.go",
		"internal/meta/languagereadiness/languagesemantic/model.go",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/toolchainconformance/contract.go",
		"internal/meta/languagereadiness/toolchainconformance/corpus.go",
		"internal/meta/languagereadiness/toolchainconformance/evaluate_test.go",
		"internal/verify/scope_language_readiness_foundation_seed.go",
	}
}
