package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"testing"
)

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
