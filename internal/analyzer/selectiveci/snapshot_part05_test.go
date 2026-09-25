package selectiveci

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"testing"
)

func TestDiffStaleSnapshotFallsBackWithoutPartialIDs(t *testing.T) {
	base := testSnapshot(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	head := testSnapshot(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	head.Sources[0].BlobDigest = testDigest("tampered")
	got, err := Diff(base, head)
	if err == nil {
		t.Fatal("Diff accepted stale head snapshot")
	}
	if got.Status != StatusUnknown || !got.FullSuiteFallback || len(got.ChangedIDs) != 0 {
		t.Fatalf("stale delta = %#v, want UNKNOWN/full-suite/no IDs", got)
	}
}
func testSnapshot(t *testing.T, path, name, id string) Snapshot {
	t.Helper()
	input := testInput(t, path, name, id)
	result, err := Build(SnapshotInput{
		Sources:         []SourceInput{input},
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{id},
	})
	if err != nil {
		t.Fatalf("Build %s: %v", path, err)
	}
	return result
}
func testInput(t *testing.T, path, name, id string) SourceInput {
	t.Helper()
	source := []byte(fmt.Sprintf("package fixture\n\n//gooo:bind id=%q role=\"HANDWRITTEN_IMPL\"\nfunc %s() {}\n", id, name))
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{
		Filename: path, PackagePath: "fixture", Source: source,
	}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semanticbinding.Extract = %#v, err=%v", result, err)
	}
	return SourceInput{Path: path, BlobDigest: testDigest(string(source)), Bindings: result.Bindings}
}
func testDigest(value string) string { return digest([]byte(value)) }
