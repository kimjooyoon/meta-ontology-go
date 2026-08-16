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
	inputDigest := snapshot.Artifacts[0].InputDigest
	snapshot.Artifacts = append(snapshot.Artifacts, Artifact{
		ID: "cache", Kind: KindCache, InputIDs: []string{"dsl"},
		InputDigest: inputDigest, ContentDigest: HashBytes([]byte("cache")),
	})
	first := Check(snapshot)
	if err := first.Error(); err != nil {
		t.Fatal(err)
	}
	if !first.Fresh() || len(first.Items) != 4 {
		t.Fatalf("unexpected fresh report: %#v", first)
	}
	reordered := snapshot
	reordered.Sources = append([]Source(nil), snapshot.Sources...)
	reordered.Artifacts = append([]Artifact(nil), snapshot.Artifacts...)
	reordered.Evidence = append([]Evidence(nil), snapshot.Evidence...)
	reverseSources(reordered.Sources)
	reverseArtifacts(reordered.Artifacts)
	reverseEvidence(reordered.Evidence)
	second := Check(reordered)
	if !reflect.DeepEqual(first.Items, second.Items) {
		t.Fatalf("check order changed result:\nfirst=%#v\nsecond=%#v", first.Items, second.Items)
	}
}

func TestCheckDoesNotMutateSnapshot(t *testing.T) {
	snapshot, _ := freshSnapshot(t)
	before := snapshot
	before.Sources = append([]Source(nil), snapshot.Sources...)
	before.Artifacts = append([]Artifact(nil), snapshot.Artifacts...)
	before.Evidence = append([]Evidence(nil), snapshot.Evidence...)
	for index := range before.Artifacts {
		before.Artifacts[index].InputIDs = append([]string(nil), snapshot.Artifacts[index].InputIDs...)
		before.Artifacts[index].EvidenceIDs = append([]string(nil), snapshot.Artifacts[index].EvidenceIDs...)
	}
	for index := range before.Evidence {
		before.Evidence[index].InputIDs = append([]string(nil), snapshot.Evidence[index].InputIDs...)
		before.Evidence[index].Provenance.UsedIDs = append([]string(nil), snapshot.Evidence[index].Provenance.UsedIDs...)
	}
	Check(snapshot)
	if !reflect.DeepEqual(snapshot, before) {
		t.Fatalf("check mutated snapshot:\nbefore=%#v\nafter=%#v", before, snapshot)
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
	assertState(t, report, KindSource, "dsl", StateFresh)
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

func TestDigestInputsIsOrderIndependentAndStrict(t *testing.T) {
	digests := map[string]string{"a": HashBytes([]byte("a")), "b": HashBytes([]byte("b"))}
	left, err := DigestInputs([]string{"b", "a"}, digests)
	if err != nil {
		t.Fatal(err)
	}
	right, err := DigestInputs([]string{"a", "b"}, digests)
	if err != nil || left != right {
		t.Fatalf("input digest was not deterministic: left=%q right=%q err=%v", left, right, err)
	}
	before := []string{"b", "a"}
	ids := append([]string(nil), before...)
	if _, err := DigestInputs(ids, digests); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, before) {
		t.Fatalf("digest calculation mutated IDs: before=%v after=%v", before, ids)
	}
	if _, err := DigestInputs([]string{"a", "a"}, map[string]string{"a": left}); err == nil {
		t.Fatal("duplicate input was accepted")
	}
	if _, err := DigestInputs([]string{""}, digests); err == nil {
		t.Fatal("empty input was accepted")
	}
	if _, err := DigestInputs([]string{"missing"}, map[string]string{}); err == nil {
		t.Fatal("missing input was accepted")
	}
	if ValidDigest("ABC" + left[3:]) {
		t.Fatal("uppercase digest was accepted")
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

func TestLoadManifestRejectsMalformedJSON(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		"null":         "null",
		"unknown":      "{\"sources\":[],\"unexpected\":true}",
		"duplicate":    "{\"sources\":[],\"sources\":[]}",
		"trailing":     "{\"sources\":[]} {}",
		"unterminated": "{\"sources\":[",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root, name+".json")
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadManifest(path); err == nil {
				t.Fatalf("malformed manifest %q was accepted", data)
			}
		})
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

func reverseSources(values []Source) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseArtifacts(values []Artifact) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseEvidence(values []Evidence) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func TestDigestInputEncodingIsStable(t *testing.T) {
	input := HashBytes([]byte("a"))
	digest, err := DigestInputs([]string{"a"}, map[string]string{"a": input})
	if err != nil {
		t.Fatal(err)
	}
	want := HashBytes([]byte("a\x00" + input + "\n"))
	if digest != want {
		t.Fatalf("unexpected input encoding digest: got %s want %s", digest, want)
	}
}
