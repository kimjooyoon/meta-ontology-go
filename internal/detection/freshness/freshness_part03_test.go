package freshness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
