package bidir

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestThreeWayDeterministicRelationShadowsCandidate(t *testing.T) {
	base := threeWayBillingModel(t)
	candidate := NewSourcedFact(CandidateFact, "billing://entity/payment", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "candidate.go", Start: 1, End: 2})
	relation := Relation{Kind: candidate.Predicate, Source: candidate.Subject, Target: candidate.Object, Span: SourceSpan{File: "accepted.go", Start: 3, End: 4}}
	left := threeWayAddRelation(t, base, relation)
	right := base.Clone()
	right.Candidates = FactSet{candidate}
	baseBefore, leftBefore, rightBefore := base.Clone(), left.Clone(), right.Clone()

	result, err := ReconcileThreeWay(base, left, right)
	if err != nil {
		t.Fatal(err)
	}
	if countRelation(result.Model, relation) != 1 || result.Model.Candidates.Contains(candidate) {
		t.Fatalf("three-way merge retained a shadowed candidate: %#v", result.Model)
	}
	if !reflect.DeepEqual(base, baseBefore) || !reflect.DeepEqual(left, leftBefore) || !reflect.DeepEqual(right, rightBefore) {
		t.Fatal("three-way candidate shadowing mutated an input")
	}
}
func TestThreeWayPartialRelationAbsenceDoesNotDelete(t *testing.T) {
	base := threeWayAddRelation(t, threeWayBillingModel(t), threeWayInvokesRelation("base.gooo", 10))
	left := base.Clone()
	left.Relations = left.Relations[:len(left.Relations)-1]
	result, err := ReconcileThreeWay(base, left, base)
	if err != nil {
		t.Fatal(err)
	}
	if !SemanticEquivalent(result.Model, base) || !result.Delta.IsEmpty() {
		t.Fatalf("partial relation absence deleted base state: %#v", result)
	}
}
func TestThreeWayConflictOrderingIsStable(t *testing.T) {
	base := threeWayBillingModel(t)
	left, right := base.Clone(), base.Clone()
	left.Nodes[0].Attributes = map[string]string{"side": "left-a"}
	left.Nodes[1].Attributes = map[string]string{"side": "left-b"}
	right.Nodes[0].Attributes = map[string]string{"side": "right-a"}
	right.Nodes[1].Attributes = map[string]string{"side": "right-b"}
	_, err := ReconcileThreeWay(base, left, right)
	var conflictErr *ThreeWayConflictError
	if !errors.As(err, &conflictErr) || len(conflictErr.Conflicts) != 2 {
		t.Fatalf("unexpected conflict set: %v", err)
	}
	if conflictErr.Conflicts[0].Identity > conflictErr.Conflicts[1].Identity || !strings.Contains(conflictErr.Error(), string(ThreeWaySameIdentity)) {
		t.Fatalf("conflicts were not deterministically ordered: %#v", conflictErr.Conflicts)
	}
}
