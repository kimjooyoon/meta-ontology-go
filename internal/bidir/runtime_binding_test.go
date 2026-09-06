package bidir

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRuntimeBindingGetPutPreservesSourceEvidence(t *testing.T) {
	file, diagnostics := syntax.ParseFile("binding.gooo", bindingFixtureForBidir)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("binding parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.RuntimeBindings) != 1 || model.RuntimeBindings[0].Entity != ID("billing://entity/payment") {
		t.Fatalf("runtime bindings=%#v", model.RuntimeBindings)
	}
	written, err := Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	if !DocumentEquivalent(document, written) {
		t.Fatalf("Get-Put changed binding source evidence: before=%#v after=%#v", document.RuntimeBindings, written.RuntimeBindings)
	}
	observed, err := Get(written)
	if err != nil || !SemanticEquivalent(model, observed) {
		t.Fatalf("Get-Put semantic binding changed: err=%v observed=%#v", err, observed.RuntimeBindings)
	}
}

func TestRuntimeBindingFanoutSurvivesBidirAndIRLowering(t *testing.T) {
	source := `package billing
namespace billing

entity Integer id "billing://entity/integer"
activity Produce(Integer) -> Integer computes "int.add:1"
activity ConsumeA(Integer) -> Integer computes "int.add:1"
activity ConsumeB(Integer) -> Integer computes "int.add:1"

bind Produce.result -> ConsumeA.input
bind Produce.result -> ConsumeB.input
`
	file, diagnostics := syntax.ParseFile("fanout.gooo", source)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("fanout parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.RuntimeBindings) != 2 || model.RuntimeBindings[0].Entity != ID("billing://entity/integer") || model.RuntimeBindings[1].Entity != ID("billing://entity/integer") {
		t.Fatalf("bidir fanout bindings=%#v", model.RuntimeBindings)
	}
	ir, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.RuntimeBindings) != 2 {
		t.Fatalf("IR fanout bindings=%#v", ir.RuntimeBindings)
	}
	for _, binding := range ir.RuntimeBindings {
		if binding.Entity != semantic.ID("billing://entity/integer") {
			t.Fatalf("IR fanout entity=%q, want billing://entity/integer", binding.Entity)
		}
	}
}

func TestRuntimeBindingMissingRelationFailsClosedWithoutPanic(t *testing.T) {
	file, diagnostics := syntax.ParseFile("binding.gooo", bindingFixtureForBidir)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("binding parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	filtered := model.Relations[:0]
	for _, relation := range model.Relations {
		if relation.Kind == PredicateUsed && relation.Source == ID("billing://activity/consume") {
			continue
		}
		filtered = append(filtered, relation)
	}
	model.Relations = filtered
	if err := model.Validate(); err == nil || !errors.Is(err, semantic.ErrRuntimeBindingPort) {
		t.Fatalf("missing binding relation error=%v, want runtime binding port error", err)
	}
}

func TestRuntimeBindingRepeatedInputArityFailsClosedAfterRelationDeduplication(t *testing.T) {
	source := `package billing
namespace billing

entity Integer id "billing://entity/integer"
entity Result id "billing://entity/result"
activity Produce() -> Integer
activity Consume(Integer, Integer) -> Result

bind Produce.result -> Consume.input
`
	file, diagnostics := syntax.ParseFile("repeated-input.gooo", source)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("repeated input parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Get(document); err == nil || !errors.Is(err, semantic.ErrRuntimeBindingPort) {
		t.Fatalf("Get accepted repeated bound input: %v", err)
	}
	if _, err := LowerDocument(document); err == nil || !errors.Is(err, semantic.ErrRuntimeBindingPort) {
		t.Fatalf("LowerDocument accepted repeated bound input: %v", err)
	}
}

func TestPutRejectsRuntimeBindingChangesWithoutWriting(t *testing.T) {
	file, diagnostics := syntax.ParseFile("binding.gooo", bindingFixtureForBidir)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("binding parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	changed := model.Clone()
	changed.RuntimeBindings[0].Entity = ID("billing://entity/order")
	written, err := Put(document, changed)
	putErr, ok := err.(*PutError)
	if !ok || putErr.Code != PutWriteConflict || !putErr.NoWrite {
		t.Fatalf("binding mutation error=%T %v", err, err)
	}
	if !reflect.DeepEqual(written, document) {
		t.Fatalf("binding mutation changed source document: got=%#v want=%#v", written, document)
	}
}

const bindingFixtureForBidir = `package billing
namespace billing

entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity Produce(Order) -> Payment
activity Consume(Payment) -> Order

bind Produce.result -> Consume.input
`
