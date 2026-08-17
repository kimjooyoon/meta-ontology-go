package eligibilityregistry

import (
	"bytes"
	"reflect"
	"testing"
)

func smokeEntries() []RegistryEntry {
	return []RegistryEntry{
		{
			ID:                 "billing://entity/order",
			Kind:               ItemSemantic,
			Authority:          AuthorityBusinessDSL,
			RequiredProjection: ProjectionSemanticIR,
		},
		{
			ID:                 "billing://entity/order",
			Kind:               ItemStructural,
			Authority:          AuthoritySemanticIR,
			RequiredProjection: ProjectionGeneratedGo,
		},
	}
}

func TestCanonicalRegistrySmoke(t *testing.T) {
	source := Digest("sha256:fecc95c83dddc9d0c844b92935769212b3ea0dd6691f062799564df8a9671c4f")
	entries := smokeEntries()
	before := append([]RegistryEntry(nil), entries...)
	frame, ok := CanonicalRegistry(source, entries)
	if !ok || len(frame) != 632 {
		t.Fatalf("frame: ok=%v length=%d", ok, len(frame))
	}
	if len(canonicalItemFrame(entries[0])) != 210 || len(canonicalItemFrame(entries[1])) != 212 {
		t.Fatal("item frame lengths")
	}
	digest, ok := RegistryDigest(source, entries)
	if !ok || digest != "sha256:b8f4d3eacd0c51aa9e4db8748466a8c394704dbc821852a6bf903607f9cd955a" {
		t.Fatalf("digest: ok=%v value=%q", ok, digest)
	}
	if !reflect.DeepEqual(entries, before) {
		t.Fatal("entries mutated")
	}
	reversed := []RegistryEntry{entries[1], entries[0]}
	reversedBefore := append([]RegistryEntry(nil), reversed...)
	reversedFrame, ok := CanonicalRegistry(source, reversed)
	if !ok || !bytes.Equal(reversedFrame, frame) {
		t.Fatal("reversed frame")
	}
	reversedDigest, ok := RegistryDigest(source, reversed)
	if !ok || reversedDigest != digest {
		t.Fatal("reversed digest")
	}
	if !reflect.DeepEqual(reversed, reversedBefore) {
		t.Fatal("reversed entries mutated")
	}
}

func TestCanonicalRegistryRejectsInvalidSource(t *testing.T) {
	frame, ok := CanonicalRegistry("sha256:0", smokeEntries())
	if ok || frame != nil {
		t.Fatal("invalid source accepted")
	}
}

func TestCanonicalRegistryRejectsInvalidEntry(t *testing.T) {
	entries := smokeEntries()
	entries[0].ID = "Order"
	if frame, ok := CanonicalRegistry(validSmokeSource(), entries); ok || frame != nil {
		t.Fatal("invalid entry accepted")
	}
}

func TestCanonicalRegistryRejectsInvalidCombination(t *testing.T) {
	entries := smokeEntries()
	entries[0].Authority = AuthoritySemanticIR
	if frame, ok := CanonicalRegistry(validSmokeSource(), entries); ok || frame != nil {
		t.Fatal("invalid combination accepted")
	}
}

func TestCanonicalRegistryRejectsDuplicate(t *testing.T) {
	entries := smokeEntries()
	entries = append(entries, entries[0])
	if frame, ok := CanonicalRegistry(validSmokeSource(), entries); ok || frame != nil {
		t.Fatal("duplicate accepted")
	}
}

func TestCanonicalRegistryRejectsConflict(t *testing.T) {
	entries := smokeEntries()
	conflict := entries[0]
	conflict.RequiredProjection = ProjectionGeneratedGo
	entries = append(entries, conflict)
	if frame, ok := CanonicalRegistry(validSmokeSource(), entries); ok || frame != nil {
		t.Fatal("conflict accepted")
	}
}

func validSmokeSource() Digest {
	return "sha256:fecc95c83dddc9d0c844b92935769212b3ea0dd6691f062799564df8a9671c4f"
}
