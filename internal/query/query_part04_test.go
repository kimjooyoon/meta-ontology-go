package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestFromSemanticIRKeepsCandidatesOutOfAuthoritativeQueries(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	order, err := semantic.NewEntity("billing://entity/order", "billing", "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(order); err != nil {
		t.Fatal(err)
	}
	invoice, err := semantic.NewEntity("billing://entity/invoice", "billing", "Invoice")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(invoice); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(semantic.NewUsedFact(activity.ID, order.ID)); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddCandidate(semantic.NewCandidateFact(activity.ID, semantic.Used, order.ID, "ambiguous adapter")); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddCandidate(semantic.NewCandidateFact(activity.ID, semantic.Used, invoice.ID, "ambiguous invoice adapter")); err != nil {
		t.Fatal(err)
	}
	projected, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	result, err := projected.ExactMatch(NewExactQuery(ID(activity.ID.String()), Used, ID(order.ID.String())))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Deterministic) != 1 || len(result.Candidates) != 0 {
		t.Fatalf("candidate was not shadowed by authoritative fact: %#v", result)
	}
	candidateResult, err := projected.ExactMatch(NewExactQuery(ID(activity.ID.String()), Used, ID(invoice.ID.String())))
	if err != nil {
		t.Fatal(err)
	}
	if len(candidateResult.Deterministic) != 0 || len(candidateResult.Candidates) != 1 || candidateResult.Candidates[0].Status != Candidate {
		t.Fatalf("candidate projection was not kept separate: %#v", candidateResult)
	}
	if projected.StableHash() == "" || projected.Canonical() == "" {
		t.Fatal("query projection did not expose a stable read fingerprint")
	}
	if result.Metadata.SemanticDigest != ir.StableHash() || result.Metadata.ProjectionStatus != "derived" {
		t.Fatalf("query result lost projection metadata: %#v", result.Metadata)
	}
	if node, ok := projected.Node(ID(activity.ID.String())); !ok || node.Kind != ActivityNodeKind {
		t.Fatalf("activity node type was not projected: %#v %t", node, ok)
	}
}
