package verify

func init() {
	branchScopeAllowlist["agent/luna-meta-24-denominator-evolution"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/denominator-evolution.yml",
		"cmd/denominator-evolution-witness",
		"cmd/denominator-evolution-verify",
		"examples/denominator-evolution",
		"internal/meta/denominatorevolution",
		"internal/meta/denominatorevolutionverify",
		"internal/verify/scope_denominator_evolution.go",
		"scripts/denominator-evolution",
	}
}
