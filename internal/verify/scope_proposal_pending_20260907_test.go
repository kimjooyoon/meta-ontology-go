package verify

import (
	"reflect"
	"testing"
)

func TestProposalPendingObservationScopeIsExact(t *testing.T) {
	want := []string{
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
	got, known := BranchScope("agent/proposal-pending-observation-20260907")
	if !known || !reflect.DeepEqual(got, want) {
		t.Fatalf("pending observation scope changed: known=%t paths=%v", known, got)
	}
	if err := CheckPathScopeForBranch(want, "agent/proposal-pending-observation-20260907"); err != nil {
		t.Fatal(err)
	}
	for _, denied := range []string{".github/workflows/ci.yml", ".github/foundation-authorization.json",
		".github/workflows/gooo-release-publish.yml", "go.mod", "internal/meta/semantic/graph.go"} {
		if err := CheckPathScopeForBranch([]string{denied}, "agent/proposal-pending-observation-20260907"); err == nil {
			t.Errorf("allowed unrelated path %q", denied)
		}
	}
}
