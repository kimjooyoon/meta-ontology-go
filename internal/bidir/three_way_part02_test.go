package bidir

import (
	"reflect"
	"testing"
)

func TestThreeWayRejectsDeleteVsModify(t *testing.T) {
	base := threeWayBillingModel(t)
	node := base.Nodes[0]
	left, err := base.Apply(Delta{RemovedNodes: []Node{node}})
	if err != nil {
		t.Fatal(err)
	}
	right := base.Clone()
	right.Nodes[0].Attributes = map[string]string{"owner": "right"}
	_, err = ReconcileThreeWay(base, left, right)
	if conflict := threeWayConflict(t, err); conflict.Code != ThreeWayDeleteVsModify {
		t.Fatalf("unexpected delete/modify conflict: %#v", conflict)
	}
}
func TestThreeWayRejectsIncompatibleRelationAttributes(t *testing.T) {
	base := threeWayAddRelation(t, threeWayBillingModel(t), threeWayInvokesRelation("base.gooo", 10))
	left, right := base.Clone(), base.Clone()
	left.Relations[0].Attributes = map[string]string{"mode": "left"}
	right.Relations[0].Attributes = map[string]string{"mode": "right"}
	_, err := ReconcileThreeWay(base, left, right)
	if conflict := threeWayConflict(t, err); conflict.Code != ThreeWayRelationAttributes || conflict.Scope != "relation" {
		t.Fatalf("unexpected relation conflict: %#v", conflict)
	}
}
func TestThreeWayRejectsInvalidEndpointWithoutMutation(t *testing.T) {
	base := threeWayBillingModel(t)
	left := base.Clone()
	invalid := Relation{Kind: PredicateInvokes, Source: "billing://activity/pay-order", Target: "billing://entity/missing"}
	invalid.ID = StableRelationID(invalid.Kind, invalid.Source, invalid.Target)
	left.Relations = append(left.Relations, invalid)
	before := left.Clone()
	result, err := ReconcileThreeWay(base, left, base)
	if conflict := threeWayConflict(t, err); conflict.Code != ThreeWayEndpointInvalid {
		t.Fatalf("unexpected endpoint conflict: %#v", conflict)
	}
	if !reflect.DeepEqual(left, before) || !SemanticEquivalent(result.Model, base) {
		t.Fatal("invalid endpoint merge mutated an input or output")
	}
}
func TestThreeWayCandidatesStaySeparateWhenObservationIsAbsent(t *testing.T) {
	base := threeWayBillingModel(t)
	candidate := NewSourcedFact(CandidateFact, "billing://entity/payment", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "candidate.go", Start: 1, End: 2})
	base.Candidates = FactSet{candidate}
	left, right := base.Clone(), base.Clone()
	left.Candidates = nil
	result, err := ReconcileThreeWay(base, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Model.Candidates.Contains(candidate) || len(result.Model.Relations) != len(base.Relations) {
		t.Fatalf("candidate was deleted or promoted: %#v", result.Model)
	}
	for _, relation := range result.Model.Relations {
		if relation.Kind == candidate.Predicate && relation.Source == candidate.Subject && relation.Target == candidate.Object {
			t.Fatal("candidate became an authoritative relation")
		}
	}
}
