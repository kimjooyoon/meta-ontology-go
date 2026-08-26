package selectiveci

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestBuildAndCanonicalJSONAreOrderIndependent(t *testing.T) {
	first := testInput(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	second := testInput(t, "pkg/customer.go", "Customer", "urn:gooo:entity:customer")

	left, err := Build(SnapshotInput{
		Sources:         []SourceInput{first, second},
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{first.Bindings[0].ID, second.Bindings[0].ID},
	})
	if err != nil {
		t.Fatalf("Build left: %v", err)
	}
	right, err := Build(SnapshotInput{
		Sources:         []SourceInput{second, first},
		SourceMapDigest: testDigest("source-map"),
		RegistryDigest:  testDigest("registry"),
		RegisteredIDs:   []string{second.Bindings[0].ID, first.Bindings[0].ID},
	})
	if err != nil {
		t.Fatalf("Build right: %v", err)
	}
	leftJSON, err := left.CanonicalJSON()
	if err != nil {
		t.Fatalf("left canonical JSON: %v", err)
	}
	rightJSON, err := right.CanonicalJSON()
	if err != nil {
		t.Fatalf("right canonical JSON: %v", err)
	}
	if left.Digest != right.Digest || !bytes.Equal(leftJSON, rightJSON) {
		t.Fatalf("input permutation changed snapshot: %s/%s", left.Digest, right.Digest)
	}
	decoded, err := DecodeSnapshot(leftJSON)
	if err != nil {
		t.Fatalf("DecodeSnapshot: %v", err)
	}
	if decoded.Digest != left.Digest {
		t.Fatalf("decoded digest = %q, want %q", decoded.Digest, left.Digest)
	}
	var roundTrip Snapshot
	if err := json.Unmarshal(leftJSON, &roundTrip); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if roundTrip.Digest != left.Digest {
		t.Fatalf("round-trip digest = %q, want %q", roundTrip.Digest, left.Digest)
	}
}
