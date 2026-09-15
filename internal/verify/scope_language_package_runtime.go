package verify

func init() {
	branchScopeAllowlist["agent/language-package-runtime"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-package-runtime-witness",
		"cmd/language-readiness-witness",
		"docs/language/language-package-runtime-readiness.md",
		"docs/language/language-package-runtime.md",
		"examples/language-package-runtime",
		"internal/packageruntime",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact/build_evidence.go",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/language_package_runtime_test.go",
		"internal/meta/languagereadiness/languagepackageruntime",
		"internal/meta/languagereadiness/package_runtime_evidence.go",
		"internal/meta/languagereadiness/promotion.go",
		"internal/verify/scope_language_package_runtime.go",
	}
}
