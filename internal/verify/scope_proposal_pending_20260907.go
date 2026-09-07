package verify

func init() {
	branchScopeAllowlist["agent/proposal-pending-observation-20260907"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/metric-counterfactual.yml",
		"docs/proposal-pending-observation.md",
		"internal/meta/metricstrategy/proposalpredecessor/collect.go",
		"internal/meta/metricstrategy/proposalpredecessor/pending_collect.go",
		"internal/meta/metricstrategy/proposalpredecessor/pending_collect_test.go",
		"internal/meta/metricstrategy/proposalpredecessor/pending_identity.go",
		"internal/meta/metricstrategy/proposalpredecessor/pending_priority_test.go",
		"internal/meta/metricstrategy/proposalpredecessor/pending_select.go",
		"internal/meta/metricstrategy/proposalpredecessor/types.go",
		"internal/verify/scope_proposal_pending_20260907.go",
		"internal/verify/scope_proposal_pending_20260907_test.go",
		"scripts/metric-strategy/main.go",
		"scripts/metric-strategy/predecessor.go",
		"scripts/metric-strategy/predecessor_observe.go",
		"scripts/metric-strategy/predecessor_wait.go",
		"scripts/metric-strategy/predecessor_wait_fixture_test.go",
		"scripts/metric-strategy/predecessor_wait_test.go",
	}
}
