package bidir

import (
	"errors"
	"reflect"
	"strings"
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

func TestThreeWayReplayIsPermutationStableAndNoWrite(t *testing.T) {
	base := threeWayBillingModel(t)
	left := threeWayAddRelation(t, base, threeWayInvokesRelation("left.gooo", 10))
	right := base.Clone()
	right.Nodes[1].Attributes = map[string]string{"tier": "gold"}
	baseBefore, leftBefore, rightBefore := base.Clone(), left.Clone(), right.Clone()
	first, err := ReconcileThreeWay(base, left, right)
	if err != nil {
		t.Fatal(err)
	}
	permutedBase, permutedLeft, permutedRight := base.Clone(), left.Clone(), right.Clone()
	reverseModelCollections(&permutedBase)
	reverseModelCollections(&permutedLeft)
	reverseModelCollections(&permutedRight)
	second, err := ReconcileThreeWay(permutedBase, permutedLeft, permutedRight)
	if err != nil {
		t.Fatal(err)
	}
	if !SemanticEquivalent(first.Model, second.Model) || !reflect.DeepEqual(first.Delta, second.Delta) || !reflect.DeepEqual(first.Locality, second.Locality) {
		t.Fatalf("replay changed merge output: first=%#v second=%#v", first, second)
	}
	if !reflect.DeepEqual(base, baseBefore) || !reflect.DeepEqual(left, leftBefore) || !reflect.DeepEqual(right, rightBefore) {
		t.Fatal("three-way merge mutated an input")
	}
}

func threeWayBillingModel(t *testing.T) Model {
	t.Helper()
	model, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func threeWayAddRelation(t *testing.T, base Model, relation Relation) Model {
	t.Helper()
	model, err := base.Apply(Delta{AddedRelations: []Relation{relation}})
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func threeWayInvokesRelation(file string, start int) Relation {
	return Relation{Kind: PredicateInvokes, Source: "billing://activity/pay-order", Target: "billing://activity/audit-payment", Span: SourceSpan{File: file, Start: start, End: start + 2}}
}

func threeWayConflict(t *testing.T, err error) ThreeWayConflict {
	t.Helper()
	if err == nil {
		t.Fatal("expected three-way conflict")
	}
	var conflictErr *ThreeWayConflictError
	if !errors.As(err, &conflictErr) || len(conflictErr.Conflicts) != 1 {
		t.Fatalf("unexpected conflict error: %T %v", err, err)
	}
	if strings.TrimSpace(conflictErr.Error()) == "" {
		t.Fatal("conflict error has no stable text")
	}
	return conflictErr.Conflicts[0]
}

func countRelation(model Model, want Relation) int {
	count := 0
	for _, relation := range model.Relations {
		if relationKey(relation.Kind, relation.Source, relation.Target) == relationKey(want.Kind, want.Source, want.Target) {
			count++
		}
	}
	return count
}

func reverseModelCollections(model *Model) {
	for left, right := 0, len(model.Nodes)-1; left < right; left, right = left+1, right-1 {
		model.Nodes[left], model.Nodes[right] = model.Nodes[right], model.Nodes[left]
	}
	for left, right := 0, len(model.Relations)-1; left < right; left, right = left+1, right-1 {
		model.Relations[left], model.Relations[right] = model.Relations[right], model.Relations[left]
	}
}
