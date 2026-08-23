package verify

func init() {
	branchScopeAllowlist["agent/language-diagnostic-provenance"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-diagnostic-provenance-readiness-binding",
		"cmd/language-diagnostic-provenance-witness",
		"cmd/language-readiness-witness",
		"docs/language/language-diagnostic-provenance-readiness-binding.md",
		"docs/language/language-diagnostic-provenance.md",
		"examples/language-diagnostic-provenance",
		"internal/meta/languageconcept",
		"internal/meta/languagereadiness/artifact",
		"internal/meta/languagereadiness/external_evidence.go",
		"internal/meta/languagereadiness/language_diagnostic_provenance_test.go",
		"internal/meta/languagereadiness/languagediagnosticprovenance",
		"internal/meta/languagereadiness/languagediagnosticprovenancebinding",
		"internal/meta/languagereadiness/languagegointeroperationbinding",
		"internal/meta/languagereadiness/promotion.go",
		"internal/verify/scope_language_diagnostic_provenance.go",
	}
}
