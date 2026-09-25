package verify

func init() {
	branchScopeAllowlist["agent/language-source-execution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/gooo",
		"cmd/language-delivery-scorecard",
		"cmd/language-source-execution-witness",
		"docs/language/language-delivery-scorecard.md",
		"docs/language/language-source-execution.md",
		"examples/language-delivery-scorecard",
		"examples/language-source-execution",
		"examples/user-journey-scorecard",
		"internal/meta/languageconcept",
		"internal/meta/languagedelivery",
		"internal/meta/languagereadiness/toolchaincli",
		"internal/meta/languagesourceexecution",
		"internal/meta/userjourneyscorecard",
		"internal/sourceexecution",
		"internal/verify/scope_language_source_execution.go",
		"scripts/language-delivery-scorecard",
		"scripts/language-source-execution",
		"scripts/user-journey-profile",
	}
}
