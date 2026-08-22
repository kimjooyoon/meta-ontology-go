package verify

func init() {
	branchScopeAllowlist["agent/concept-governed-refactoring-v29"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"examples/language-concept-catalog/README.md",
		"internal/meta/languageconcept",
		"internal/meta/metricstrategy",
		"internal/meta/metricstrategy/verify",
		"internal/verify/scope_concept_strategy.go",
		"scripts/metric-strategy",
	}
}
