package verify

import (
	"bytes"
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"go/format"
	"strings"
	"testing"
)

func TestSemanticRoundTripEvidenceAndScope(t *testing.T) {
	document := verificationDocument()
	if err := bidir.CheckGetPut(document); err != nil {
		t.Fatalf("Get-Put: %v", err)
	}
	model, err := bidir.Get(document)
	if err != nil {
		t.Fatal(err)
	}
	fact := bidir.NewSourcedFact(bidir.DeterministicFact, "billing://activity/pay-order", bidir.PredicateInvokes, "billing://activity/audit-payment", bidir.SourceSpan{File: "payment.go", Start: 10, End: 20})
	fact.SubjectKind = bidir.ActivityKind
	fact.ObjectKind = bidir.ActivityKind
	result, err := bidir.Reconcile(model, bidir.FactDelta{Added: bidir.FactSet{fact}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Delta.IsEmpty() || !result.Locality.Contains(fact.Subject) || !result.Locality.Contains(fact.Object) || result.Locality.Contains("billing://entity/unrelated") {
		t.Fatalf("semantic scope was not deterministic: delta=%#v locality=%#v", result.Delta, result.Locality)
	}
	candidate := bidir.NewSourcedFact(bidir.CandidateFact, fact.Subject, bidir.PredicateWasDerivedFrom, "billing://entity/order", bidir.SourceSpan{File: "payment.go", Start: 21, End: 30})
	result, err = bidir.Reconcile(model, bidir.FactDelta{Added: bidir.FactSet{candidate}})
	if err != nil || !result.Model.Candidates.Contains(candidate) || !bidir.SemanticEquivalent(model, result.Model) {
		t.Fatalf("candidate evidence crossed semantic boundary: err=%v result=%#v", err, result)
	}
	withoutSource := bidir.NewFact(bidir.DeterministicFact, fact.Subject, bidir.PredicateInvokes, fact.Object)
	_, err = bidir.Reconcile(model, bidir.FactDelta{Added: bidir.FactSet{withoutSource}})
	var reconcileErr *bidir.ReconcileError
	if !errors.As(err, &reconcileErr) || reconcileErr.Conflicts[0].Kind != bidir.ConflictMissingSource {
		t.Fatalf("missing evidence was accepted: %v", err)
	}
}
func TestGeneratedFreshnessAndFormatting(t *testing.T) {
	ir := generator.SemanticIR{
		Package:  "billinggen",
		Entities: []generator.Entity{{ID: "billing://entity/order", Name: "Order", GoName: "Order"}},
		Activities: []generator.Activity{{
			ID: "billing://activity/pay-order", Name: "PayOrder", GoName: "PayOrder",
			Inputs: []generator.Port{{ID: "billing://entity/order", Name: "order", GoName: "order", EntityID: "billing://entity/order", GoType: "Order"}},
		}},
	}
	first, err := generator.Generate(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := generator.Generate(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Source, second.Source) {
		t.Fatal("generated output changed between identical runs")
	}
	formatted, err := format.Source(first.Source)
	if err != nil || !bytes.Equal(formatted, first.Source) {
		t.Fatalf("generated output is not gofmt-stable: %v", err)
	}
	for _, marker := range []string{"//gooo:generated:start", "//gooo:generated:end", "billing://activity/pay-order"} {
		if !strings.Contains(string(first.Source), marker) {
			t.Fatalf("generated output is missing %q", marker)
		}
	}
}
