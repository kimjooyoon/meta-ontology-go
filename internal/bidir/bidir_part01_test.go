package bidir

import (
	"testing"
)

func billingDocument() Document {
	return Document{
		Package:   "billing",
		Namespace: "billing",
		Declarations: []Declaration{
			{Kind: EntityKind, ID: "billing://entity/order", Name: "Order"},
			{Kind: EntityKind, ID: "billing://entity/payment", Name: "Payment"},
			{Kind: EntityKind, ID: "billing://entity/audit", Name: "Audit"},
			{Kind: EntityKind, ID: "billing://entity/unrelated", Name: "Unrelated"},
			{Kind: ActivityKind, Name: "PayOrder", Inputs: []Reference{{Name: "Order"}}, Outputs: []Reference{{Name: "Payment"}}},
			{Kind: ActivityKind, Name: "AuditPayment", Inputs: []Reference{{Name: "Payment"}}, Outputs: []Reference{{Name: "Audit"}}},
		},
	}
}
func TestGenericGetPutLaw(t *testing.T) {
	document := billingDocument()
	if err := CheckGetPut(document); err != nil {
		t.Fatal(err)
	}
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	written, err := Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	if len(written.Declarations) != len(document.Declarations) {
		t.Fatalf("write-back changed declaration count: got %d want %d", len(written.Declarations), len(document.Declarations))
	}
}
func TestAcceptedGoDeltaIsVisibleAfterPutGet(t *testing.T) {
	document := billingDocument()
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "payment.go", Start: 42, End: 58})
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = ActivityKind
	reconciled, err := Reconcile(base, FactDelta{Added: FactSet{fact}})
	if err != nil {
		t.Fatal(err)
	}
	if reconciled.Delta.IsEmpty() || len(reconciled.Delta.AddedRelations) != 1 {
		t.Fatalf("expected one semantic relation delta: %#v", reconciled.Delta)
	}
	updatedDocument, err := Put(document, reconciled.Model)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := Get(updatedDocument)
	if err != nil {
		t.Fatal(err)
	}
	if !SemanticEquivalent(reconciled.Model, observed) {
		t.Fatalf("accepted Go delta was not visible after Put-Get\nwant: %s\ngot:  %s", SemanticFingerprint(reconciled.Model), SemanticFingerprint(observed))
	}
	if err := CheckPutGet(document, reconciled.Model); err != nil {
		t.Fatal(err)
	}
}
