package query

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestProjectionMetadataBindsAuthorityAndKnownEmptyEvidence(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}

	projected, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	metadata := projected.Metadata()
	if metadata.SchemaVersion != QueryProjectionSchemaVersion {
		t.Fatalf("schema version = %q", metadata.SchemaVersion)
	}
	if metadata.ProjectionStatus != "derived" {
		t.Fatalf("projection status = %q", metadata.ProjectionStatus)
	}
	if metadata.GraphHash != projected.StableHash() {
		t.Fatalf("graph hash is not bound to current view: %q != %q", metadata.GraphHash, projected.StableHash())
	}
	if metadata.SemanticDigest != ir.StableHash() {
		t.Fatalf("semantic digest = %q, want %q", metadata.SemanticDigest, ir.StableHash())
	}
	if metadata.SourceDigest != "" || metadata.SourceStatus != "unavailable" {
		t.Fatalf("source was fabricated: %#v", metadata)
	}
	if metadata.EvidenceStatus != "known_empty" || metadata.ProvenanceStatus != "known_empty" {
		t.Fatalf("empty evidence was not typed explicitly: %#v", metadata)
	}
	if metadata.EvidenceDigest == "" || metadata.ProvenanceDigest == "" {
		t.Fatalf("validated IR digests missing: %#v", metadata)
	}
	labels := make(map[string]AuthorityLabel, len(metadata.AuthorityLabels))
	for _, label := range metadata.AuthorityLabels {
		labels[label.View] = label
	}
	if labels["semantic_ir"].Authority != "authoritative" || labels["semantic_ir"].Status != "bound" {
		t.Fatalf("semantic authority label = %#v", labels["semantic_ir"])
	}
	if labels["query_graph"].Authority != "derived" || labels["query_graph"].Status != "current" {
		t.Fatalf("query authority label = %#v", labels["query_graph"])
	}
	if labels[".gooo"].Status != "unavailable" {
		t.Fatalf("source authority status = %#v", labels[".gooo"])
	}

	metadata.AuthorityLabels[0].Status = "tampered"
	if projected.Metadata().AuthorityLabels[0].Status == "tampered" {
		t.Fatal("metadata labels leaked mutable internal state")
	}
}

func TestUnboundGraphMetadataDoesNotClaimAuthority(t *testing.T) {
	metadata := New().Metadata()
	if metadata.ProjectionStatus != "unbound" || metadata.SemanticDigest != "" {
		t.Fatalf("unbound graph claimed an IR binding: %#v", metadata)
	}
	for _, label := range metadata.AuthorityLabels {
		if label.View == "query_graph" && (label.Authority != "derived" || label.Status != "unbound") {
			t.Fatalf("unbound graph authority label = %#v", label)
		}
	}
}
