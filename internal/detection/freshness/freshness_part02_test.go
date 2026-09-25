package freshness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckDetectsMissingRecordsAndFiles(t *testing.T) {
	snapshot, root := freshSnapshot(t)
	snapshot.RequiredArtifacts = []Requirement{{ID: "missing-cache", Kind: KindCache}}
	snapshot.RequiredEvidence = []Requirement{{ID: "missing-proof", Kind: KindEvidence}}
	if err := os.Remove(filepath.Join(root, "semantic.go")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "proof.json")); err != nil {
		t.Fatal(err)
	}
	report := Check(snapshot)
	assertState(t, report, KindProjection, "projection", StateMissing)
	assertState(t, report, KindCache, "missing-cache", StateMissing)
	assertState(t, report, KindEvidence, "proof", StateMissing)
	assertState(t, report, KindEvidence, "missing-proof", StateMissing)
}
func TestCheckRejectsInvalidProvenanceReferencesAndKinds(t *testing.T) {
	snapshot, _ := freshSnapshot(t)
	snapshot.Evidence[0].Provenance.ActivityID = ""
	snapshot.Artifacts[0].Provenance.UsedIDs = []string{"unknown"}
	snapshot.Artifacts = append(snapshot.Artifacts, Artifact{
		ID: "bad-kind", Kind: "other", InputDigest: snapshot.Artifacts[0].InputDigest,
		ContentDigest: HashBytes([]byte("bad-kind")),
	})
	snapshot.RequiredEvidence = []Requirement{{ID: "wrong", Kind: KindProjection}}
	report := Check(snapshot)
	assertState(t, report, KindEvidence, "proof", StateInvalid)
	assertState(t, report, KindProjection, "projection", StateInvalid)
	assertState(t, report, Kind("other"), "bad-kind", StateInvalid)
	assertState(t, report, KindProjection, "wrong", StateInvalid)
}
func TestCheckRejectsDuplicateIDs(t *testing.T) {
	snapshot, _ := freshSnapshot(t)
	snapshot.Artifacts = append(snapshot.Artifacts, Artifact{
		ID: "projection", Kind: KindCache, InputDigest: snapshot.Artifacts[0].InputDigest,
		ContentDigest: HashBytes([]byte("duplicate")),
	})
	report := Check(snapshot)
	found := false
	for _, item := range report.Items {
		if item.ID == "projection" && item.State == StateInvalid {
			found = true
		}
	}
	if !found {
		t.Fatalf("duplicate ID was not invalid: %#v", report.Items)
	}
}
