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
	records, reason := normalizeEntries(smokeEntries())
	if reason != ReasonNone {
		t.Fatal("valid records rejected")
	}
	if canonicalRegistryFrame("sha256:0", records) != nil {
		t.Fatal("private registry frame accepted malformed source")
	}
	invalidID := smokeEntries()
	invalidID[0].ID = "Order"
	if canonicalItemFrame(invalidID[0]) != nil {
		t.Fatal("private item frame accepted invalid ID")
	}
	frame, ok = CanonicalRegistry(validSmokeSource(), invalidID)
	if ok || frame != nil {
		t.Fatal("invalid ID frame accepted")
	}
	digest, ok = RegistryDigest(validSmokeSource(), invalidID)
	if ok || digest != "" {
		t.Fatal("invalid ID digest accepted")
	}
	invalidCombination := smokeEntries()
	invalidCombination = append(invalidCombination, invalidCombination[0])
	invalidCombination[2].RequiredProjection = ProjectionGeneratedGo
	frame, ok = CanonicalRegistry(validSmokeSource(), invalidCombination)
	if ok || frame != nil {
		t.Fatal("same-key altered field frame accepted")
	}
	digest, ok = RegistryDigest(validSmokeSource(), invalidCombination)
	if ok || digest != "" {
		t.Fatal("same-key altered field digest accepted")
	}
	duplicate := append(smokeEntries(), smokeEntries()[0])
	frame, ok = CanonicalRegistry(validSmokeSource(), duplicate)
	if ok || frame != nil {
		t.Fatal("duplicate frame accepted")
	}
	digest, ok = RegistryDigest(validSmokeSource(), duplicate)
	if ok || digest != "" {
		t.Fatal("duplicate digest accepted")
	}
}
