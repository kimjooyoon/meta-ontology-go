package bidir

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

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
