package verify

func init() {
	branchScopeAllowlist["agent/promotion-authorized-continuity"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/promotion-authorized-continuity",
		"docs/language/promotion-authorized-continuity.md",
		"internal/meta/languagereadiness/promotioncontinuity",
		"internal/verify/scope_promotion_authorized_continuity.go",
	}
	branchScopeAllowlist["agent/promotion-continuity-v1"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"cmd/promotion-authorized-continuity",
		"docs/language/promotion-authorized-continuity.md",
		"examples/metric-meta-program/main.gooo",
		"internal/meta/languagereadiness/promotioncontinuity",
		"internal/meta/languagereadiness/languagesyntax/conformance/evaluate_test.go",
		"internal/meta/metricprogram",
		"internal/meta/metricstrategy",
		"internal/verify/scope_promotion_authorized_continuity.go",
	}
	branchScopeAllowlist["agent/promotion-continuity-recovery-v1"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/meta/languagereadiness/promotioncontinuity",
		"internal/verify/scope_promotion_authorized_continuity.go",
	}
}
