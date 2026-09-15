package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func TestLowerDerivesProvFacts(t *testing.T) {
	file, diagnostics := syntax.Parse(`package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment`)
	if diagnostics.Error() != nil {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	ir, err := Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Graph.Nodes()) != 3 || len(ir.Graph.Facts()) != 2 {
		t.Fatalf("unexpected graph: %#v", ir)
	}
	if !ir.Graph.HasFact(semantic.FactKey{Subject: "billing://activity/pay-order", Predicate: semantic.Used, Object: "billing://entity/order"}) {
		t.Fatal("missing used fact")
	}
}
func TestSyntaxAdapterAndDocumentLowererAgree(t *testing.T) {
	file, diagnostics := syntax.ParseFile("billing.gooo", `package billing
namespace billing
entity Order id "billing://entity/order"
entity Payment id "billing://entity/payment"
activity PayOrder(Order) -> Payment`)
	if diagnostics.Error() != nil {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	fromSyntax, err := Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	fromDocument, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	if !EquivalentAfterRoundTrip(fromSyntax, fromDocument) {
		t.Fatalf("syntax and parser-neutral lowerers disagree:\n%s\n%s", fromSyntax.SemanticCanonical(), fromDocument.SemanticCanonical())
	}
	if document.Declarations[0].Span.File != "billing.gooo" {
		t.Fatalf("declaration source span was not adapted: %#v", document.Declarations[0].Span)
	}
}
