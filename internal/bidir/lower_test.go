package bidir

import (
	"reflect"
	"sort"
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

func TestTypedLowererRetainsOutputPortSpansAndOrder(t *testing.T) {
	document := sourceOrderedOutputDocument()
	first, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	firstOutputs := typedOutputFacts(first)
	secondOutputs := typedOutputFacts(second)
	if !reflect.DeepEqual(firstOutputs, secondOutputs) {
		t.Fatalf("repeated typed lowering changed output evidence: %#v != %#v", firstOutputs, secondOutputs)
	}
	wantIDs := []semantic.ID{"billing://entity/zebra", "billing://entity/apple"}
	if got := outputFactIDsBySpan(firstOutputs); !reflect.DeepEqual(got, wantIDs) {
		t.Fatalf("typed lowering lost authoritative output order: got %v want %v", got, wantIDs)
	}
	if len(firstOutputs) != 2 || firstOutputs[0].Span == firstOutputs[1].Span {
		t.Fatalf("typed lowering did not retain two distinct output spans: %#v", firstOutputs)
	}
}

func typedOutputFacts(ir semantic.IR) []semantic.Fact {
	activity := semantic.MustIdentity("billing://activity/process")
	var outputs []semantic.Fact
	for _, fact := range ir.Graph.Facts() {
		if fact.Predicate == semantic.WasGeneratedBy && fact.Object == activity {
			outputs = append(outputs, fact)
		}
	}
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].Span.Start.Offset < outputs[j].Span.Start.Offset })
	return outputs
}

func outputFactIDsBySpan(facts []semantic.Fact) []semantic.ID {
	ids := make([]semantic.ID, len(facts))
	for index, fact := range facts {
		ids[index] = fact.Subject
	}
	return ids
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
