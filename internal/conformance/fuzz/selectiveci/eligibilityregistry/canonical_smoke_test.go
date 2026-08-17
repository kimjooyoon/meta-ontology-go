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

func validSmokeSource() Digest {
	return "sha256:fecc95c83dddc9d0c844b92935769212b3ea0dd6691f062799564df8a9671c4f"
}

func TestCanonicalRegistrySmoke(t *testing.T) {
	entries := smokeEntries()
	frame, ok := CanonicalRegistry(validSmokeSource(), entries)
	if !ok || len(frame) != 632 {
		t.Fatalf("frame: ok=%v length=%d", ok, len(frame))
	}
	if len(canonicalItemFrame(entries[0])) != 210 || len(canonicalItemFrame(entries[1])) != 212 {
		t.Fatal("item frame lengths")
	}
	digest, ok := RegistryDigest(validSmokeSource(), entries)
	if !ok || digest != "sha256:b8f4d3eacd0c51aa9e4db8748466a8c394704dbc821852a6bf903607f9cd955a" {
		t.Fatalf("digest: ok=%v value=%q", ok, digest)
	}
	if !reflect.DeepEqual(entries, smokeEntries()) {
		t.Fatal("entries mutated")
	}
	reversed := []RegistryEntry{entries[1], entries[0]}
	reversedFrame, ok := CanonicalRegistry(validSmokeSource(), reversed)
	if !ok || !bytes.Equal(reversedFrame, frame) {
		t.Fatal("reversed frame")
	}
	reversedDigest, ok := RegistryDigest(validSmokeSource(), reversed)
	if !ok || reversedDigest != digest {
		t.Fatal("reversed digest")
	}
	if !reflect.DeepEqual(reversed, []RegistryEntry{entries[1], entries[0]}) {
		t.Fatal("reversed entries mutated")
	}
}

func TestCanonicalRegistryRejections(t *testing.T) {
	frame, ok := CanonicalRegistry("sha256:0", smokeEntries())
	if ok || frame != nil {
		t.Fatal("malformed source frame accepted")
	}
	digest, ok := RegistryDigest("sha256:0", smokeEntries())
	if ok || digest != "" {
		t.Fatal("malformed source digest accepted")
	}
	invalidID := smokeEntries()
	invalidID[0].ID = "Order"
	invalidCombination := smokeEntries()
	invalidCombination[0].Authority = AuthoritySemanticIR
	duplicate := smokeEntries()
	duplicate = append(duplicate, duplicate[0])
	conflict := smokeEntries()
	conflictingEntry := conflict[0]
	conflictingEntry.RequiredProjection = ProjectionGeneratedGo
	conflict = append(conflict, conflictingEntry)
	tests := []struct {
		name    string
		entries []RegistryEntry
	}{
		{
			name:    "invalid ID",
			entries: invalidID,
		},
		{
			name:    "invalid combination",
			entries: invalidCombination,
		},
		{
			name:    "duplicate",
			entries: duplicate,
		},
		{
			name:    "conflict",
			entries: conflict,
		},
	}
	for _, test := range tests {
		frame, ok := CanonicalRegistry(validSmokeSource(), test.entries)
		if ok || frame != nil {
			t.Fatalf("%s frame accepted", test.name)
		}
		digest, ok := RegistryDigest(validSmokeSource(), test.entries)
		if ok || digest != "" {
			t.Fatalf("%s digest accepted", test.name)
		}
	}
}
