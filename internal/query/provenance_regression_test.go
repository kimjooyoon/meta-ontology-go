package query

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func evidenceFixture(t *testing.T, reverse bool) semantic.IR {
	t.Helper()
	ir := semantic.NewIR("fixture", "fixture")
	activity, err := semantic.NewActivity("fixture://activity/compile", "fixture", "Compile")
	if err != nil {
		t.Fatal(err)
	}
	source, err := semantic.NewEntity("fixture://entity/source", "fixture", "Source")
	if err != nil {
		t.Fatal(err)
	}
	output, err := semantic.NewEntity("fixture://entity/output", "fixture", "Output")
	if err != nil {
		t.Fatal(err)
	}
	nodes := []semantic.Node{activity, source, output}
	facts := []semantic.Fact{
		semantic.NewUsedFact(activity.ID, source.ID),
		semantic.NewWasGeneratedByFact(output.ID, activity.ID),
	}
	if reverse {
		nodes[0], nodes[2] = nodes[2], nodes[0]
		facts[0], facts[1] = facts[1], facts[0]
	}
	for _, node := range nodes {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	digest := semantic.StableHashString("fixture payload/v1")
	for _, fact := range facts {
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
		evidence, err := semantic.NewEvidence(
			semantic.MustIdentity("fixture://evidence/"+string(fact.Predicate)),
			semantic.GoHostedCompilerID, semantic.CompilerRunEvidence, fact.Key(), digest,
		)
		if err != nil {
			t.Fatal(err)
		}
		if err := ir.AddEvidence(evidence); err != nil {
			t.Fatal(err)
		}
	}
	return ir
}

func projectionLabel(metadata ProjectionMetadata, view string) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == view {
			return label
		}
	}
	return AuthorityLabel{}
}
