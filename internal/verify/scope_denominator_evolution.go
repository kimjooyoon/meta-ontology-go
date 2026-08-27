package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-24-denominator-evolution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/denominator-evolution.yml",
		".github/workflows/experiment-promotion.yml",
		"cmd/denominator-evolution-witness",
		"cmd/denominator-evolution-verify",
		"cmd/experiment-promotion-witness",
		"cmd/experiment-promotion-verify",
		"examples/denominator-evolution",
		"examples/experiment-promotion",
		"internal/meta/denominatorevolution",
		"internal/meta/denominatorevolutionverify",
		"internal/meta/experimentpromotion",
		"internal/meta/experimentpromotionverify",
		"internal/verify/scope_denominator_evolution.go",
		"scripts/denominator-evolution",
		"scripts/experiment-promotion",
	}
}
