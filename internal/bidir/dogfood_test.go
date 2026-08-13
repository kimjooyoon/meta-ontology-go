package bidir

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestBillingGoooDogfoodRoundTripsDSLGoAndBack(t *testing.T) {
	document := billingGoooDocument(t)
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckGetPut(document); err != nil {
		t.Fatalf("Get-Put: %v", err)
	}

	projected := ProjectFacts(base)
	for _, fact := range projected {
		if !fact.Source.Valid() {
			t.Fatalf("projected fact lost source provenance: %#v", fact)
		}
	}
	lifted, err := LiftFacts(base, projected)
	if err != nil {
		t.Fatalf("project/lift: %v", err)
	}
	if !SemanticEquivalent(base, lifted.Model) {
		t.Fatalf("project/lift changed meaning: %s != %s", SemanticFingerprint(base), SemanticFingerprint(lifted.Model))
	}

	derived := NewSourcedFact(
		DeterministicFact,
		"billing://entity/payment",
		PredicateWasDerivedFrom,
		"billing://entity/order",
		SourceSpan{File: "examples/billing/handwritten.go", Start: 6, End: 24},
	)
	derived.SubjectKind = EntityKind
	derived.ObjectKind = EntityKind
	updated, err := Reconcile(base, FactDelta{Added: FactSet{derived}})
	if err != nil {
		t.Fatalf("lift Go fact: %v", err)
	}
	if len(updated.Delta.AddedRelations) != 1 || !updated.Locality.Contains(derived.Subject) || !updated.Locality.Contains(derived.Object) {
		t.Fatalf("Go fact delta was not localized: delta=%#v locality=%#v", updated.Delta, updated.Locality)
	}

	written, err := Put(document, updated.Model)
	if err != nil {
		t.Fatalf("Put accepted Go fact: %v", err)
	}
	observed, err := Get(written)
	if err != nil {
		t.Fatalf("Get written document: %v", err)
	}
	if !SemanticEquivalent(updated.Model, observed) {
		t.Fatalf("DSL -> Go -> DSL changed meaning: %s != %s", SemanticFingerprint(updated.Model), SemanticFingerprint(observed))
	}
	relation, found := findRelation(observed, PredicateWasDerivedFrom, derived.Subject, derived.Object)
	if !found || relation.Span != derived.Source {
		t.Fatalf("accepted provenance was not preserved: found=%v relation=%#v", found, relation)
	}
}

func TestPutRejectsInvalidSourceWithoutWrite(t *testing.T) {
	document := billingDocument()
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	invalid := document
	invalid.Relations = append(invalid.Relations, Relation{
		Kind:   PredicateWasDerivedFrom,
		Source: "billing://entity/missing",
		Target: "billing://entity/order",
	})

	written, err := Put(invalid, model)
	if err == nil {
		t.Fatal("Put accepted an invalid source document")
	}
	if !reflect.DeepEqual(written, invalid) {
		t.Fatalf("invalid source was not returned unchanged: %#v", written)
	}
	assertPutNoWrite(t, err, PutSourceInvalid)
}

func TestPutRejectsMissingSpanWithoutWrite(t *testing.T) {
	document := billingDocument()
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	updated := base.Clone()
	updated.Relations = append(updated.Relations, Relation{
		Kind:   PredicateWasDerivedFrom,
		Source: "billing://entity/payment",
		Target: "billing://entity/order",
	})

	written, err := Put(document, updated)
	if err == nil {
		t.Fatal("Put accepted a semantic update without provenance")
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatalf("missing-span update changed the source view: %#v", written)
	}
	assertPutNoWrite(t, err, PutProvenanceMissing)
}

func TestPutRejectsConflictingRelationWithoutWrite(t *testing.T) {
	document := billingDocument()
	conflicting := document
	conflicting.Relations = []Relation{
		{Kind: PredicateWasDerivedFrom, Source: "billing://entity/payment", Target: "billing://entity/order"},
		{Kind: PredicateWasDerivedFrom, Source: "billing://entity/payment", Target: "billing://entity/order", Attributes: map[string]string{"source": "conflict"}},
	}
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}

	written, err := Put(conflicting, base)
	if err == nil {
		t.Fatal("Put accepted conflicting source relations")
	}
	if !reflect.DeepEqual(written, conflicting) {
		t.Fatalf("conflicting source was not returned unchanged: %#v", written)
	}
	assertPutNoWrite(t, err, PutSourceInvalid)
}

func TestReconcileRejectsInvalidEvidenceTransactionally(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	bad := NewSourcedFact(
		DeterministicFact,
		"billing://activity/pay-order",
		PredicateInvokes,
		"billing://activity/audit-payment",
		SourceSpan{File: "payment.go", Start: 20, End: 10},
	)

	result, err := Reconcile(base, FactDelta{Added: FactSet{bad}})
	if err == nil {
		t.Fatal("invalid source span was accepted")
	}
	var reconcileErr *ReconcileError
	if !errors.As(err, &reconcileErr) || len(reconcileErr.Conflicts) != 1 || reconcileErr.Conflicts[0].Kind != ConflictInvalidFact {
		t.Fatalf("unexpected invalid evidence error: %v", err)
	}
	if !SemanticEquivalent(base, result.Model) || !result.Delta.IsEmpty() {
		t.Fatalf("invalid evidence changed the model: result=%#v", result)
	}
}

func assertPutNoWrite(t *testing.T, err error, code string) {
	t.Helper()
	var putErr *PutError
	if !errors.As(err, &putErr) || !putErr.NoWrite || putErr.Code != code {
		t.Fatalf("unexpected Put rejection: %v", err)
	}
}

func billingGoooDocument(t *testing.T) Document {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locating dogfood test failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
	path := filepath.Join(root, "examples", "billing", "main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.Error() != nil {
		t.Fatalf("parse %s: %v", path, diagnostics.Error())
	}
	if _, err := Lower(file); err != nil {
		t.Fatalf("lower %s to semantic IR: %v", path, err)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatalf("adapt %s: %v", path, err)
	}
	return document
}
