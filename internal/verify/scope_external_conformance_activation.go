package verify

func init() {
	branchScopeAllowlist["agent/external-assurance-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/external-conformance-activation-witness",
		"docs/language/language-semantic-readiness-binding.md",
		"docs/language/language-semantic-model.md",
		"docs/language/language-syntax-roundtrip.md",
		"docs/language/toolchain-conformance.md",
		"examples/external-conformance-activation",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languageassurance/digest.go",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/externalconformanceactivation",
		"internal/meta/languageassurance/verticalsliceclosureshadow",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/registry_accessors.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/meta/languagereadiness/languagesyntax",
		"examples/language-semantic-model/corpus.json",
		"examples/language-semantic-model/README.md",
		"examples/toolchain-conformance/corpus.json",
		"examples/vertical-slice-closure-shadow/README.md",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/meta/languagereadiness/languagesemanticbinding",
		"internal/meta/languagereadiness/toolchainconformance",
		"internal/verify/scope_external_conformance_activation.go",
	}
}
