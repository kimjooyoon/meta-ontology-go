package bidir

import (
	"testing"

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

const bindingFixtureForBidir = `package billing
namespace billing

entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity Produce(Order) -> Payment
activity Consume(Payment) -> Order

bind Produce.result -> Consume.input
`
