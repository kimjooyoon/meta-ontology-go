package bidir

import (
	"errors"
	"reflect"
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

func TestExplicitRemovalIsAFirstClassDelta(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateUsed, "billing://entity/order", SourceSpan{File: "payment.go", Start: 10, End: 20})
	result, err := Reconcile(base, FactDelta{Removed: FactSet{fact}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.RemovedRelations) != 1 {
		t.Fatalf("expected removed relation: %#v", result.Delta)
	}
	if _, found := findRelation(result.Model, PredicateUsed, fact.Subject, fact.Object); found {
		t.Fatal("removed relation still present")
	}
	// Repeating an already-applied removal is idempotent and does not need a
	// second source span because it changes no semantic state.
	second, err := Reconcile(result.Model, FactDelta{Removed: FactSet{NewFact(DeterministicFact, fact.Subject, fact.Predicate, fact.Object)}})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Delta.IsEmpty() {
		t.Fatalf("idempotent removal changed the model: %#v", second.Delta)
	}
}

func TestSemanticEquivalenceIgnoresPresentationAndFingerprintIsStable(t *testing.T) {
	left, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	right := left.Clone()
	right.Nodes[0].Name = "Renamed for display"
	right.Nodes[0].Namespace = "other-display-namespace"
	right.Nodes[0].Aliases = []string{"legacy", "new"}
	right.Nodes[0].Span = SourceSpan{File: "renamed.gooo", Start: 1, End: 8}
	if !SemanticEquivalent(left, right) {
		t.Fatal("presentation-only changes broke semantic equivalence")
	}
	if SemanticFingerprint(left) != SemanticFingerprint(right) {
		t.Fatal("presentation-only changes changed semantic fingerprint")
	}
	right.Nodes[0].Attributes = map[string]string{"effect": "changed"}
	if SemanticEquivalent(left, right) {
		t.Fatal("semantic attribute change was ignored")
	}
}

func TestLocalityExcludesUnrelatedNodes(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "payment.go", Start: 30, End: 40})
	updated, err := Reconcile(base, FactDelta{Added: FactSet{fact}})
	if err != nil {
		t.Fatal(err)
	}
	locality := updated.Locality
	if !locality.Contains("billing://activity/pay-order") || !locality.Contains("billing://activity/audit-payment") {
		t.Fatalf("changed endpoints absent from locality: %#v", locality)
	}
	if locality.Contains("billing://entity/unrelated") {
		t.Fatalf("unrelated node included in locality: %#v", locality)
	}
	unchanged := LocalityBetween(base, base)
	if len(unchanged.Touched) != 0 || len(unchanged.Affected) != 0 {
		t.Fatalf("implementation-only/no semantic edit should be local to nothing: %#v", unchanged)
	}
}

func TestDiffApplyPreservesSemanticDeltaAndIgnoresPresentation(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	updated := base.Clone()
	updated.Nodes[0].Name = "Display Rename"
	updated.Relations = append(updated.Relations, Relation{
		Kind:   PredicateInvokes,
		Source: "billing://activity/pay-order",
		Target: "billing://activity/audit-payment",
	})
	delta := Diff(base, updated)
	if len(delta.AddedRelations) != 1 || len(delta.AddedNodes) != 0 || len(delta.RemovedNodes) != 0 {
		t.Fatalf("unexpected semantic delta: %#v", delta)
	}
	applied, err := base.Apply(delta)
	if err != nil {
		t.Fatal(err)
	}
	if !SemanticEquivalent(updated, applied) {
		t.Fatalf("delta application changed meaning: %s != %s", SemanticFingerprint(updated), SemanticFingerprint(applied))
	}
	presentationOnly := base.Clone()
	presentationOnly.Nodes[0].Name = "Another Rename"
	if delta := Diff(base, presentationOnly); !delta.IsEmpty() {
		t.Fatalf("presentation-only edit produced semantic delta: %#v", delta)
	}
}

func TestFactSetNormalizationIsDeterministic(t *testing.T) {
	one := NewSourcedFact(DeterministicFact, "b", PredicateInvokes, "c", SourceSpan{File: "b.go", Start: 1, End: 2})
	two := NewSourcedFact(DeterministicFact, "a", PredicateInvokes, "c", SourceSpan{File: "a.go", Start: 1, End: 2})
	set := FactSet{one, two, one}
	want := FactSet{two, one}
	if got := set.Normalized(); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized facts:\ngot  %#v\nwant %#v", got, want)
	}
}
