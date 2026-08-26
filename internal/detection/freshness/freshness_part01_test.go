package freshness

import (
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
