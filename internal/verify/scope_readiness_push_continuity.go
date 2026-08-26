package verify

func init() {
	branchScopeAllowlist["agent/readiness-push-continuity"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"docs/language/readiness-push-continuity.md",
		"internal/verify/readiness_push_continuity_test.go",
		"internal/verify/scope_readiness_push_continuity.go",
	}
}
