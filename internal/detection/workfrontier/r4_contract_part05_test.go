package workfrontier

import (
	"reflect"
	"testing"
)

func TestR4GovernedProjectionMutationsRequireRebinding(t *testing.T) {
	state := r4FixtureInput(t, "acyclic")
	state.States[0].Status = "PASS"
	assertR4BindingMismatch(t, state, R4ReasonSnapshotBindingMismatch)

	path := r4FixtureInput(t, "acyclic")
	path.Paths[0].PolicyPriority++
	assertR4BindingMismatch(t, path, R4ReasonSnapshotBindingMismatch)

	root := r4FixtureInput(t, "acyclic")
	root.RootObligationIDs = []string{"obligation/child"}
	assertR4BindingMismatch(t, root, R4ReasonSnapshotBindingMismatch)

	policy := r4FixtureInput(t, "acyclic")
	policy.Capacity.CPUCoreNS++
	assertR4BindingMismatch(t, policy, R4ReasonPolicyBindingMismatch)

	bound := r4FixtureInput(t, "self-loop")
	bound.Rules[0].MaxIterations++
	assertR4BindingMismatch(t, bound, R4ReasonPolicyBindingMismatch)

	use := r4FixtureInput(t, "self-loop")
	use.Rules[0].IterationsUsed++
	assertR4BindingMismatch(t, use, R4ReasonSnapshotBindingMismatch)

	registry := r4FixtureInput(t, "acyclic")
	registry.Pressures[0].StableID = "pressure/changed"
	assertR4BindingMismatch(t, registry, R4ReasonRegistryBindingMismatch)
}
func assertR4BindingMismatch(t *testing.T, input R4Input, reason string) {
	t.Helper()
	got := EvaluateR4(input)
	if got.Status != R4StatusUnknown || got.Reason != reason || len(got.SelectedIDs) != 0 {
		t.Fatalf("binding mismatch = %#v, want UNKNOWN/%s with empty selection", got, reason)
	}
}
func TestR4ReceiptIsComponentwiseAndExact(t *testing.T) {
	acyclic := EvaluateR4(r4FixtureInput(t, "acyclic"))
	wantAcyclic := R4WorkReceipt{
		GraphNodes: 2, GraphEdges: 1, ReachableNodes: 2, ReachableEdges: 1,
		SCCs: 2, CyclicSCCs: 0, CondensationEdges: 1, RuleChecks: 2,
		IterationChecks: 0, PathChecks: 2, ConflictChecks: 0,
	}
	if !reflect.DeepEqual(acyclic.WorkReceipt, wantAcyclic) {
		t.Fatalf("acyclic receipt = %#v, want %#v", acyclic.WorkReceipt, wantAcyclic)
	}
	cycle := EvaluateR4(r4FixtureInput(t, "self-loop"))
	wantCycle := R4WorkReceipt{
		GraphNodes: 1, GraphEdges: 1, ReachableNodes: 1, ReachableEdges: 1,
		SCCs: 1, CyclicSCCs: 1, CondensationEdges: 0, RuleChecks: 1,
		IterationChecks: 1, PathChecks: 1, ConflictChecks: 0,
	}
	if !reflect.DeepEqual(cycle.WorkReceipt, wantCycle) {
		t.Fatalf("cycle receipt = %#v, want %#v", cycle.WorkReceipt, wantCycle)
	}
}
