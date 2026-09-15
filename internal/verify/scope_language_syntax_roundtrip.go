package verify

func init() {
	branchScopeAllowlist["agent/language-syntax-roundtrip"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"cmd/language-syntax-witness",
		"docs/language/language-syntax-roundtrip.md",
		"examples/language-concept-catalog/README.md",
		"examples/language-syntax-roundtrip",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact/build_evidence.go",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/language_syntax_test.go",
		"internal/meta/languagereadiness/languagesyntax",
		"internal/meta/languagereadiness/promotion.go",
		"internal/verify/scope_language_syntax_roundtrip.go",
	}
}
