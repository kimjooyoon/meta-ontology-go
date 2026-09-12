package verify

func init() {
	branchScopeAllowlist["agent/concept-operation-binding-receipt-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		".github/workflows/transformation-effect.yml",
		"cmd/language-readiness-witness",
		"internal/meta/languagereadiness",
		"internal/meta/metricprogram",
		"internal/meta/metricstrategy",
		"internal/verify/scope_concept_operation_binding_receipt.go",
		"scripts/metric-meta-program",
	}
}
