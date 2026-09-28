package query

import (
	"testing"

	provenance "github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestStoryWithEvidenceBindingRejectsMismatchedEvidence(t *testing.T) {
	ir := evidenceFixture(t, false)
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	target := graph.Nodes()[0].ID
	metadata := graph.Metadata()
	story, err := graph.StoryWithEvidenceBinding(target, provenance.Snapshot{
		Digest: "sha256:ledger-binding",
		Records: []provenance.Evidence{{
			ID: "evidence://stale", SemanticID: target.String(), Producer: "ci",
			Kind: provenance.KindVerification, Status: provenance.StatusVerified,
			SourceDigest: "sha256:source", SemanticDigest: "sha256:stale",
			GraphDigest: metadata.GraphHash, Sequence: 1, Hash: "sha256:event-stale",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := story.Validate(); err != nil {
		t.Fatal(err)
	}
	if story.Status != StoryUnknown || story.Reason != StoryMismatchedBindingReason {
		t.Fatalf("stale evidence was promoted: %#v", story)
	}
}

func TestStoryWithEvidenceBindingRejectsMissingSourceIdentity(t *testing.T) {
	ir := evidenceFixture(t, false)
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	target := graph.Nodes()[0].ID
	metadata := graph.Metadata()
	story, err := graph.StoryWithEvidenceBinding(target, provenance.Snapshot{
		Digest: "sha256:ledger-binding",
		Records: []provenance.Evidence{{
			ID: "evidence://unbound", SemanticID: target.String(), Producer: "ci",
			Kind: provenance.KindVerification, Status: provenance.StatusVerified,
			SemanticDigest: metadata.SemanticDigest, GraphDigest: metadata.GraphHash,
			Sequence: 1, Hash: "sha256:event-unbound",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if story.Status != StoryUnknown || story.Reason != StoryMismatchedBindingReason {
		t.Fatalf("source-less evidence was promoted: %#v", story)
	}
}
