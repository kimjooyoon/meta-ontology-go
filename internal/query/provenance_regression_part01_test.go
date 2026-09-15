package query

import (
	"reflect"
	"testing"
)

func TestEvidenceDoesNotFabricateProvenanceAcrossPermutations(t *testing.T) {
	left, right := evidenceFixture(t, false), evidenceFixture(t, true)
	leftGraph, err := FromSemanticIR(left)
	if err != nil {
		t.Fatal(err)
	}
	rightGraph, err := FromSemanticIR(right)
	if err != nil {
		t.Fatal(err)
	}
	leftMetadata, rightMetadata := leftGraph.Metadata(), rightGraph.Metadata()
	if leftMetadata.EvidenceStatus != "available" || rightMetadata.EvidenceStatus != "available" {
		t.Fatalf("evidence status was lost: %#v %#v", leftMetadata, rightMetadata)
	}
	for _, metadata := range []ProjectionMetadata{leftMetadata, rightMetadata} {
		if metadata.ProvenanceStatus != "unknown" || metadata.ProvenanceDigest != "" {
			t.Fatalf("evidence fabricated provenance: %#v", metadata)
		}
		if label := projectionLabel(metadata, "provenance"); label.Status != "unknown" {
			t.Fatalf("provenance authority was promoted: %#v", label)
		}
	}
	if leftMetadata.SemanticDigest != rightMetadata.SemanticDigest ||
		leftMetadata.EvidenceDigest != rightMetadata.EvidenceDigest {
		t.Fatalf("permuted metadata changed: %#v %#v", leftMetadata, rightMetadata)
	}
	request := traversalEnvelope(ID("fixture://activity/compile"), LayerAll, 2, 10)
	leftResponse, err := leftGraph.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	rightResponse, err := rightGraph.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	leftJSON, err := leftResponse.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	rightJSON, err := rightResponse.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(leftJSON, rightJSON) || leftResponse.Hash != rightResponse.Hash {
		t.Fatalf("permuted response changed: %s/%s vs %s/%s", leftJSON, leftResponse.Hash, rightJSON, rightResponse.Hash)
	}
	if leftResponse.Metadata.ProvenanceStatus != StatusDeferred {
		t.Fatalf("missing provenance was not deferred in envelope: %#v", leftResponse.Metadata)
	}
}
