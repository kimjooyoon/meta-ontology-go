package verify

func init() {
	branchScopeAllowlist["agent/toolchain-conformance"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"cmd/toolchain-conformance-witness",
		"docs/language/toolchain-conformance-readiness.md",
		"docs/language/toolchain-conformance.md",
		"examples/language-concept-catalog/README.md",
		"examples/toolchain-conformance",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/toolchain_conformance_evidence.go",
		"internal/meta/languagereadiness/toolchain_conformance_promotion.go",
		"internal/meta/languagereadiness/toolchain_conformance_test.go",
		"internal/meta/languagereadiness/toolchainconformance",
		"internal/verify/scope_toolchain_conformance.go",
	}
}
