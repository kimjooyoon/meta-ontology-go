package verify

func init() {
	branchScopeAllowlist["agent/language-delivery-scorecard"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/language-delivery-scorecard",
		"docs/language/language-delivery-scorecard.md",
		"examples/language-delivery-scorecard",
		"internal/meta/languageconcept",
		"internal/meta/languagedelivery",
		"internal/verify/scope_language_delivery_scorecard.go",
		"scripts/language-delivery-scorecard",
	}
}
