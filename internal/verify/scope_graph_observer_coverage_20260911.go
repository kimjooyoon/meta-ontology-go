package verify

func init() {
	branchScopeAllowlist["agent/graph-observer-coverage-20260911"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/ci.yml",
		"internal/verify/scope_graph_observer_coverage_20260911.go",
		"internal/verify/workflow_part03_test.go",
		"scripts/ci-proof/domain.go",
		"scripts/ci-proof/graph_observation.go",
		"scripts/ci-proof/graph_observation_test.go",
	}
}
