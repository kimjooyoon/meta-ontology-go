package query

import (
	"testing"

	provenance "github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestStoryBindsStableIdentityRelationshipsAndLedgerEvidence(t *testing.T) {
	ir := evidenceFixture(t, false)
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	nodes := graph.Nodes()
	if len(nodes) == 0 {
		t.Fatal("fixture has no semantic nodes")
	}
	target := nodes[0].ID
	metadata := graph.Metadata()
	snapshot := provenance.Snapshot{
		Digest: "sha256:ledger-story",
		Records: []provenance.Evidence{
			{
				ID: "evidence://verified", SemanticID: target.String(), Producer: "ci",
				Kind: provenance.KindVerification, Status: provenance.StatusVerified,
				SourceDigest: "sha256:source", SemanticDigest: metadata.SemanticDigest,
				GraphDigest: metadata.GraphHash, Sequence: 1, Hash: "sha256:event-1",
				Attributes: map[string]string{"run_id": "native-1"},
			},
			{
				ID: "evidence://rejected", SemanticID: target.String(), Producer: "ci",
				Kind: provenance.KindObservation, Status: provenance.StatusRejected,
				SourceDigest: "sha256:source", SemanticDigest: metadata.SemanticDigest,
				GraphDigest: metadata.GraphHash, Sequence: 2, Hash: "sha256:event-2",
				Attributes: map[string]string{"run_id": "native-2"},
			},
		},
	}
	story, err := graph.Story(target, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := story.Validate(); err != nil {
		t.Fatal(err)
	}
	if story.Status != StoryContradictory || story.Reason != "VERIFIED_AND_REJECTED_EVIDENCE" ||
		story.LedgerDigest != snapshot.Digest || len(story.Evidence) != 2 ||
		len(story.Relationships) == 0 || story.Evidence[0].Attributes["run_id"] != "native-1" {
		t.Fatalf("incomplete story response: %#v", story)
	}
	tampered := story
	tampered.Evidence = append([]StoryEvidenceReference(nil), story.Evidence...)
	tampered.Evidence[0].Producer = "forged"
	if tampered.Validate() == nil {
		t.Fatal("tampered story was accepted")
	}
}

func TestStoryKeepsMissingEvidenceExplicit(t *testing.T) {
	ir := evidenceFixture(t, false)
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	target := graph.Nodes()[0].ID
	story, err := graph.Story(target, provenance.Snapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if err := story.Validate(); err != nil {
		t.Fatal(err)
	}
	if story.Status != StoryUnknown || story.Reason != "MISSING_EVIDENCE" || len(story.Relationships) == 0 {
		t.Fatalf("missing evidence was hidden: %#v", story)
	}
	unknown, err := graph.Story(ID("urn:unknown"), provenance.Snapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if unknown.Status != StoryUnknown || unknown.Reason != "UNKNOWN_SEMANTIC_ID" || len(unknown.Relationships) != 0 {
		t.Fatalf("unknown endpoint was not explicit: %#v", unknown)
	}
}
