package verify

import (
	"bytes"
	"strings"
	"testing"
)

func TestEvidenceCanonicalizationComparesGoAndGooo(t *testing.T) {
	goEvidence := sampleEvidence(EvidenceProducerGo, []EvidenceFact{
		{ID: "b", Kind: "scope", Value: "billing"},
		{ID: "a", Kind: "roundtrip", Value: "pass"},
	})
	goooEvidence := sampleEvidence(EvidenceProducerGooo, []EvidenceFact{
		{ID: "a", Kind: "roundtrip", Value: "pass"},
		{ID: "b", Kind: "scope", Value: "billing"},
	})
	goPayload, err := goEvidence.CanonicalPayload()
	if err != nil {
		t.Fatal(err)
	}
	goooPayload, err := goooEvidence.CanonicalPayload()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(goPayload, goooPayload) {
		t.Fatalf("canonical payloads differ:\n%s\n%s", goPayload, goooPayload)
	}
	if err := CompareEvidence(goEvidence, goooEvidence); err != nil {
		t.Fatal(err)
	}
	goManifest, err := goEvidence.Manifest()
	if err != nil {
		t.Fatal(err)
	}
	goooManifest, err := goooEvidence.Manifest()
	if err != nil {
		t.Fatal(err)
	}
	if goManifest.PayloadSHA256 != goooManifest.PayloadSHA256 {
		t.Fatal("equivalent evidence did not produce the same digest")
	}
}
func TestEvidenceMismatchAndValidationFailClosed(t *testing.T) {
	left := sampleEvidence(EvidenceProducerGo, []EvidenceFact{{ID: "a", Kind: "scope", Value: "billing"}})
	right := sampleEvidence(EvidenceProducerGooo, []EvidenceFact{{ID: "a", Kind: "scope", Value: "other"}})
	if err := CompareEvidence(left, right); err == nil || !strings.Contains(err.Error(), "evidence mismatch") {
		t.Fatalf("mismatched evidence was accepted: %v", err)
	}
	cases := []EvidenceArtifact{
		{Producer: "python", Bundle: left.Bundle},
		{Producer: EvidenceProducerGo, Bundle: EvidenceBundle{Schema: EvidenceSchemaVersion, Stage: 4, Fixture: "billing", Decision: "pass"}},
		{Producer: EvidenceProducerGo, Bundle: EvidenceBundle{Schema: EvidenceSchemaVersion, Stage: StageGoBaseline, Fixture: "", Decision: "pass"}},
		{Producer: EvidenceProducerGo, Bundle: EvidenceBundle{Schema: EvidenceSchemaVersion, Stage: StageGoBaseline, Fixture: "billing", Decision: "pass", Facts: []EvidenceFact{{ID: "a", Kind: "scope"}, {ID: "a", Kind: "roundtrip"}}}},
	}
	for _, evidence := range cases {
		if _, err := evidence.CanonicalPayload(); err == nil {
			t.Fatal("invalid evidence was accepted")
		}
	}
}
