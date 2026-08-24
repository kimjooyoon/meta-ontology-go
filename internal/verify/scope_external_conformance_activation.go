package verify

func init() {
	branchScopeAllowlist["agent/external-assurance-activation"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/self-improvement-contract.yml",
		"cmd/external-conformance-activation-witness",
		"examples/external-conformance-activation",
		"examples/language-syntax-roundtrip/corpus.json",
		"internal/meta/languageassurance/digest.go",
		"internal/meta/languageassurance/evaluate_test.go",
		"internal/meta/languageassurance/externalconformanceactivation",
		"internal/meta/languageassurance/registry.go",
		"internal/meta/languageassurance/registry_accessors.go",
		"internal/meta/languageassurance/source_authority_activation.go",
		"internal/meta/languagereadiness/languagesyntax",
		"examples/language-semantic-model/corpus.json",
		"examples/language-semantic-model/README.md",
		"internal/meta/languagereadiness/languagesemantic",
		"internal/verify/scope_external_conformance_activation.go",
	}
}
