package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestGeneratedResolutionViewFromSemanticIR(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	business := mustSemanticEntity(t, "billing://entity/order", "Order")
	activity := mustSemanticActivity(t, "billing://activity/pay", "PayOrder")
	generated := mustSemanticEntity(t, "billing://entity/payment-code", "PaymentCode")
	for _, node := range []semantic.Node{business, activity, generated} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := ir.AddFact(semantic.NewUsedFact(activity.ID, business.ID)); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(semantic.NewWasGeneratedByFact(generated.ID, activity.ID)); err != nil {
		t.Fatal(err)
	}
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	beforeCanonical, beforeNodes, beforeHash := graph.Canonical(), graph.Nodes(), graph.StableHash()
	request := resolutionRequest(ID(business.ID.String()), LayerDeterministic, 10)
	response, err := graph.ResolveGeneratedCode(request)
	if err != nil || response.Status != ResolutionResolved {
		t.Fatalf("resolution failed: %#v %v", response, err)
	}
	if len(response.Deterministic) != 1 || len(response.Candidates) != 0 {
		t.Fatalf("resolution layers = %#v/%#v", response.Deterministic, response.Candidates)
	}
	row := response.Deterministic[0]
	if row.Business != ID(business.ID.String()) || row.Activity != ID(activity.ID.String()) ||
		row.GeneratedEntity != ID(generated.ID.String()) || row.Depth != ResolutionMaxDepth {
		t.Fatalf("resolution row = %#v", row)
	}
	if row.Status != DerivedFactStatus || row.SourceLayer != FactDeterministic.String() {
		t.Fatalf("resolution authority = %#v", row)
	}
	if response.Metadata.SemanticDigest != ir.StableHash() ||
		response.Metadata.DerivedStatus != DerivedStatusNonAuthoritative {
		t.Fatalf("resolution metadata = %#v", response.Metadata)
	}
	if label := resolutionLabel(response.Metadata); label.Authority != "derived" || label.Status != ResolutionResolved {
		t.Fatalf("resolution label = %#v", label)
	}
	if label := resolutionLabel(response.Metadata); label.View != "resolution_view" {
		t.Fatalf("resolution label view = %#v", label)
	}
	if response.Metadata.ProvenanceStatus != StatusDeferred {
		t.Fatalf("resolution fabricated provenance = %#v", response.Metadata)
	}
	if label := authorityLabel(response.Metadata, "generated_go"); label.Status != "unavailable" {
		t.Fatalf("resolution fabricated generated Go = %#v", label)
	}
	requestHash, err := request.CanonicalDigest()
	if err != nil || response.RequestHash != requestHash {
		t.Fatalf("request digest = %q/%q", response.RequestHash, requestHash)
	}
	responseHash, err := response.CanonicalDigest()
	if err != nil || response.Hash != responseHash {
		t.Fatalf("response digest = %q/%q", response.Hash, responseHash)
	}
	if graph.Canonical() != beforeCanonical || !reflect.DeepEqual(graph.Nodes(), beforeNodes) ||
		graph.StableHash() != beforeHash {
		t.Fatal("resolution mutated graph authority")
	}
}
