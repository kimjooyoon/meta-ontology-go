package verify

func init() {
	branchScopeAllowlist["agent/user-journey-scorecard"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/transformation-effect.yml",
		"cmd/user-journey-scorecard",
		"docs/language/user-journey-scorecard.md",
		"examples/user-journey-scorecard",
		"internal/meta/userjourneyscorecard",
		"internal/verify/scope_user_journey_scorecard.go",
		"scripts/user-journey-profile",
	}
}
