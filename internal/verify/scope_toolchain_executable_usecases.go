package verify

func init() {
	branchScopeAllowlist["agent/toolchain-executable-usecases"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"cmd/toolchain-usecase-witness",
		"docs/language/toolchain-executable-use-cases.md",
		"examples/language-concept-catalog/README.md",
		"examples/toolchain-executable-use-cases",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact/build_evidence.go",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/promotion.go",
		"internal/meta/languagereadiness/toolchain_usecases_test.go",
		"internal/meta/languagereadiness/toolchainusecases",
		"internal/verify/scope_toolchain_executable_usecases.go",
	}
}
