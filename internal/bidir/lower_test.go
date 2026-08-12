package bidir

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func TestCandidateDoesNotBecomeDeterministic(t *testing.T) {
	graph := semantic.NewGraph()
	entity := semantic.MustIdentity("billing://entity/order")
	if err := graph.AddNode(semantic.Node{ID: entity, Kind: semantic.Entity, Namespace: "billing", Name: "Order"}); err != nil {
		t.Fatal(err)
	}
	candidate := semantic.NewCandidateFact(entity, semantic.WasDerivedFrom, entity, "ambiguous Go call")
	if err := graph.AddCandidate(candidate); err != nil {
		t.Fatal(err)
	}
	if graph.HasFact(candidate.Key()) {
		t.Fatal("candidate was promoted unexpectedly")
	}
}

func TestPromoteCandidateIsTransactionalAndExplicit(t *testing.T) {
	ir := semantic.NewIR("billing", semantic.Namespace("billing"))
	activity := semantic.MustIdentity("billing://activity/pay-order")
	order := semantic.MustIdentity("billing://entity/order")
	if err := ir.AddNode(semantic.Node{ID: activity, Kind: semantic.Activity, Namespace: "billing", Name: "PayOrder"}); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(semantic.Node{ID: order, Kind: semantic.Entity, Namespace: "billing", Name: "Order"}); err != nil {
		t.Fatal(err)
	}
	candidate := semantic.NewCandidateFact(order, semantic.WasDerivedFrom, order, "needs review")
	if err := ir.AddCandidate(candidate); err != nil {
		t.Fatal(err)
	}
	promotedIR, promoted, err := PromoteCandidate(ir, candidate.Key())
	if err != nil {
		t.Fatal(err)
	}
	if promoted.Status != semantic.FactDeterministic || !promotedIR.Graph.HasFact(candidate.Key()) {
		t.Fatalf("candidate was not promoted: %#v", promoted)
	}
	if promotedIR.Graph.HasCandidate(candidate.Key()) || ir.Graph.HasFact(candidate.Key()) || !ir.Graph.HasCandidate(candidate.Key()) {
		t.Fatal("promotion was not explicit and transactional")
	}
}
