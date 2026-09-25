package verify

func init() {
	branchScopeAllowlist["agent/gooo-exact-subject-transport"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-language-candidate.yml",
		".github/workflows/self-improvement-language-observation.yml",
		"examples/language-semantic-model/corpus.json",
		"examples/language-syntax-roundtrip/corpus.json",
		"examples/self-improvement/TRANSPORT.md",
		"examples/self-improvement/transport.gooo",
		"examples/toolchain-conformance/corpus.json",
		"examples/vertical-slice-closure-shadow/README.md",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/languagereadiness/languagesyntax/model.go",
		"internal/meta/languagereadiness/languagesyntax/registry.go",
		"internal/meta/languagereadiness/languagesemantic/model.go",
		"internal/meta/languagereadiness/languagesemantic/registry_definition.go",
		"internal/meta/languagereadiness/languagesemanticbinding/denominator.go",
		"internal/meta/languagereadiness/languagesemanticbinding/validate_semantic_summary_test.go",
		"internal/meta/languagereadiness/toolchainconformance/contract.go",
		"internal/meta/languagereadiness/toolchainconformance/corpus.go",
		"internal/meta/languagereadiness/toolchainconformance/evaluate_test.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/contract.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/denominator.go",
		"internal/meta/languageassurance/verticalsliceclosureshadow/evidence/denominator.json",
		"internal/meta/selfimprovementtransport",
		"internal/verify/scope_gooo_exact_subject_transport.go",
		"scripts/self-improvement-transport",
	}
}
