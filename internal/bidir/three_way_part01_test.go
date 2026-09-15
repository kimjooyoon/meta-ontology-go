package bidir

import (
	"testing"
)

func TestThreeWayPresentationChangesDoNotConflict(t *testing.T) {
	base := threeWayBillingModel(t)
	left, right := base.Clone(), base.Clone()
	left.Nodes[0].Name = "left display"
	right.Nodes[0].Name = "right display"
	left.Nodes[0].Span = SourceSpan{File: "left.gooo", Start: 2, End: 4}
	right.Nodes[0].Span = SourceSpan{File: "right.gooo", Start: 8, End: 10}
	result, err := ReconcileThreeWay(base, left, right)
	if err != nil || !result.Succeeded() {
		t.Fatalf("presentation-only merge conflicted: result=%#v err=%v", result, err)
	}
	if !SemanticEquivalent(result.Model, base) || !result.Delta.IsEmpty() || len(result.Locality.Affected) != 0 {
		t.Fatalf("presentation-only merge changed semantic state: %#v", result)
	}
}
func TestThreeWayIdenticalRelationChangeAppliesOnce(t *testing.T) {
	base := threeWayBillingModel(t)
	relation := threeWayInvokesRelation("left.gooo", 10)
	left := threeWayAddRelation(t, base, relation)
	right := threeWayAddRelation(t, base, relation)
	result, err := ReconcileThreeWay(base, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.AddedRelations) != 1 || countRelation(result.Model, relation) != 1 {
		t.Fatalf("identical relation was not applied once: %#v", result)
	}
}
func TestThreeWayDisjointChangesMergeWithLocality(t *testing.T) {
	base := threeWayBillingModel(t)
	left := base.Clone()
	left.Nodes[0].Attributes = map[string]string{"retention": "long"}
	right := threeWayAddRelation(t, base, threeWayInvokesRelation("right.gooo", 20))
	result, err := ReconcileThreeWay(base, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if result.Model.Nodes[0].Attributes["retention"] != "long" || len(result.Delta.AddedRelations) != 1 {
		t.Fatalf("disjoint changes did not merge: %#v", result)
	}
	if !result.Locality.Contains("billing://entity/order") || !result.Locality.Contains("billing://activity/pay-order") || result.Locality.Contains("billing://entity/unrelated") {
		t.Fatalf("unexpected merge locality: %#v", result.Locality)
	}
}
func TestThreeWayRejectsSameIdentityChangeWithFingerprints(t *testing.T) {
	base := threeWayBillingModel(t)
	left, right := base.Clone(), base.Clone()
	left.Nodes[0].Attributes = map[string]string{"owner": "left"}
	right.Nodes[0].Attributes = map[string]string{"owner": "right"}
	result, err := ReconcileThreeWay(base, left, right)
	conflict := threeWayConflict(t, err)
	if conflict.Code != ThreeWaySameIdentity || conflict.Scope != "node" {
		t.Fatalf("unexpected identity conflict: %#v", conflict)
	}
	if conflict.BaseFingerprint != SemanticFingerprint(base) || conflict.LeftFingerprint != SemanticFingerprint(left) || conflict.RightFingerprint != SemanticFingerprint(right) {
		t.Fatalf("conflict fingerprints are not exact: %#v", conflict)
	}
	if !SemanticEquivalent(result.Model, base) || len(result.Delta.AddedNodes) != 0 {
		t.Fatalf("conflicting merge mutated output state: %#v", result)
	}
}
