package bidir

import (
	"errors"
	"testing"
)

func TestProjectedFactsRoundTripWithoutNewDelta(t *testing.T) {
	model, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	facts := ProjectFacts(model)
	reconciled, err := Reconcile(model, FactDelta{Added: facts})
	if err != nil {
		t.Fatal(err)
	}
	if !reconciled.Delta.IsEmpty() {
		t.Fatalf("projected facts created a new delta: %#v", reconciled.Delta)
	}
	if !SemanticEquivalent(model, reconciled.Model) {
		t.Fatal("DSL -> IR -> facts -> IR changed semantic meaning")
	}
}
func TestFactLayersAndSourceBoundary(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	syntactic := NewSourcedFact(SyntacticFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "payment.go", Start: 1, End: 2})
	candidate := NewSourcedFact(CandidateFact, "billing://activity/pay-order", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "payment.go", Start: 3, End: 4})
	deterministic := NewSourcedFact(DeterministicFact, "billing://entity/payment", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "payment.go", Start: 5, End: 6})
	result, err := Reconcile(base, FactDelta{Added: FactSet{syntactic, candidate, deterministic}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Syntactic) != 1 || len(result.Candidates) != 1 || len(result.Delta.AddedRelations) != 1 {
		t.Fatalf("fact layers were not separated: %#v", result)
	}
	if result.Model.Candidates.ByLayer(CandidateFact).Contains(candidate) == false {
		t.Fatal("candidate evidence was not retained")
	}
	if SemanticEquivalent(base, result.Model) {
		t.Fatal("deterministic fact did not update semantic IR")
	}

	withoutSource := NewFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment")
	rejected, err := Reconcile(base, FactDelta{Added: FactSet{withoutSource}})
	if err == nil {
		t.Fatal("unattributed semantic addition was accepted")
	}
	var reconcileErr *ReconcileError
	if !errors.As(err, &reconcileErr) || len(reconcileErr.Conflicts) != 1 || reconcileErr.Conflicts[0].Kind != ConflictMissingSource {
		t.Fatalf("unexpected reconciliation error: %v", err)
	}
	if !SemanticEquivalent(base, rejected.Model) {
		t.Fatal("failed reconciliation was not transactional")
	}
}
