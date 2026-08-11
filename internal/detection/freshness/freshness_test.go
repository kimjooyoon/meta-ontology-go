package freshness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheckFreshSnapshotIsDeterministic(t *testing.T) {
	snapshot, _ := freshSnapshot(t)
	first := Check(snapshot)
	if err := first.Error(); err != nil {
		t.Fatal(err)
	}
	if !first.Fresh() || len(first.Items) != 3 {
		t.Fatalf("unexpected fresh report: %#v", first)
	}
	reordered := snapshot
	reordered.Sources = append([]Source(nil), snapshot.Sources...)
	reordered.Artifacts = append([]Artifact(nil), snapshot.Artifacts...)
	reordered.Evidence = append([]Evidence(nil), snapshot.Evidence...)
	second := Check(reordered)
	if !reflect.DeepEqual(first.Items, second.Items) {
		t.Fatalf("check order changed result:\nfirst=%#v\nsecond=%#v", first.Items, second.Items)
	}
}

func TestCheckDetectsStaleInputsAndContent(t *testing.T) {
	snapshot, root := freshSnapshot(t)
	if err := os.WriteFile(filepath.Join(root, "main.gooo"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "semantic.go"), []byte("changed output"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "proof.json"), []byte("changed evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := Check(snapshot)
	assertState(t, report, KindProjection, "projection", StateStale)
	assertState(t, report, KindEvidence, "proof", StateStale)
}

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

func TestCheckRejectsInvalidProvenanceAndInputReferences(t *testing.T) {
	snapshot, _ := freshSnapshot(t)
	snapshot.Evidence[0].Provenance.ActivityID = ""
	snapshot.Artifacts[0].Provenance.UsedIDs = []string{"unknown"}
	report := Check(snapshot)
	assertState(t, report, KindEvidence, "proof", StateInvalid)
	assertState(t, report, KindProjection, "projection", StateInvalid)
}

func TestDigestInputsIsOrderIndependentAndStrict(t *testing.T) {
	left, err := DigestInputs([]string{"b", "a"}, map[string]string{"a": HashBytes([]byte("a")), "b": HashBytes([]byte("b"))})
	if err != nil {
		t.Fatal(err)
	}
	right, err := DigestInputs([]string{"a", "b"}, map[string]string{"a": HashBytes([]byte("a")), "b": HashBytes([]byte("b"))})
	if err != nil || left != right {
		t.Fatalf("input digest was not deterministic: left=%q right=%q err=%v", left, right, err)
	}
	if _, err := DigestInputs([]string{"a", "a"}, map[string]string{"a": left}); err == nil {
		t.Fatal("duplicate input was accepted")
	}
	if _, err := DigestInputs([]string{"missing"}, map[string]string{}); err == nil {
		t.Fatal("missing input was accepted")
	}
}

func TestCheckManifestResolvesRelativePaths(t *testing.T) {
	snapshot, root := freshSnapshot(t)
	snapshot.Root = ""
	manifest := filepath.Join(root, "freshness.json")
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, data, 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := CheckManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Fresh() {
		t.Fatalf("manifest was not fresh: %#v", report.Items)
	}
}

func freshSnapshot(t *testing.T) (Snapshot, string) {
	t.Helper()
	root := t.TempDir()
	write := func(name, content string) string {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return name
	}
	sourcePath := write("main.gooo", "activity PayOrder")
	projectionPath := write("semantic.go", "generated projection")
	evidencePath := write("proof.json", "verification evidence")
	sourceDigest := HashBytes([]byte("activity PayOrder"))
	inputDigest, err := DigestInputs([]string{"dsl"}, map[string]string{"dsl": sourceDigest})
	if err != nil {
		t.Fatal(err)
	}
	return Snapshot{
		Root:    root,
		Sources: []Source{{ID: "dsl", Path: sourcePath}},
		Artifacts: []Artifact{{
			ID: "projection", Kind: KindProjection, Path: projectionPath,
			InputIDs: []string{"dsl"}, InputDigest: inputDigest,
			ContentDigest: HashBytes([]byte("generated projection")), EvidenceIDs: []string{"proof"},
		}},
		Evidence: []Evidence{{
			ID: "proof", Path: evidencePath, InputIDs: []string{"dsl"}, InputDigest: inputDigest,
			ContentDigest: HashBytes([]byte("verification evidence")),
			Provenance:    Provenance{ActivityID: "verify:generated", EntityID: "proof", UsedIDs: []string{"projection"}},
		}},
		RequiredArtifacts: []Requirement{{ID: "projection", Kind: KindProjection}},
		RequiredEvidence:  []Requirement{{ID: "proof", Kind: KindEvidence}},
	}, root
}

func assertState(t *testing.T, report Report, kind Kind, id string, expected State) {
	t.Helper()
	for _, item := range report.Items {
		if item.Kind == kind && item.ID == id {
			if item.State != expected {
				t.Fatalf("%s/%s state=%s, want %s: %s", kind, id, item.State, expected, item.Detail)
			}
			return
		}
	}
	t.Fatalf("missing result for %s/%s: %#v", kind, id, report.Items)
}
