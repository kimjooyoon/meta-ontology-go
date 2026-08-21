package selectiveci

import (
	"bytes"
	"encoding/json"
)

// CanonicalJSON returns the strict canonical JSON representation of a
// snapshot, including and verifying its digest.
func (s Snapshot) CanonicalJSON() ([]byte, error) {
	normalized, err := normalizeSnapshot(s)
	if err != nil {
		return nil, err
	}
	unsigned, err := normalized.unsignedJSON()
	if err != nil {
		return nil, err
	}
	if want := digest(unsigned); normalized.Digest != want {
		return nil, fail(CodeStaleSnapshot, "snapshot digest %q does not match %q", normalized.Digest, want)
	}
	return json.Marshal(wireForSnapshot(normalized))
}

// StableHash returns the digest bound into the canonical snapshot.
func (s Snapshot) StableHash() string { return s.Digest }

// Validate checks the digest-bound canonical representation without external state.
func (s Snapshot) Validate() error {
	_, err := s.CanonicalJSON()
	return err
}

// MarshalJSON makes ordinary encoding/json use the strict canonical form.
func (s Snapshot) MarshalJSON() ([]byte, error) { return s.CanonicalJSON() }

// UnmarshalJSON accepts only the exact canonical JSON representation.
func (s *Snapshot) UnmarshalJSON(data []byte) error {
	parsed, err := DecodeSnapshot(data)
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// DecodeSnapshot decodes and verifies a strict canonical JSON snapshot.
func DecodeSnapshot(data []byte) (Snapshot, error) {
	wire, err := decodeSnapshotWire(data)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot := Snapshot{
		Schema: wire.Schema, Status: wire.Status, FullSuiteFallback: wire.FullSuiteFallback,
		SourceMapDigest: wire.SourceMapDigest, RegistryDigest: wire.RegistryDigest,
		RegisteredIDs: wire.RegisteredIDs, Sources: wire.Sources, Digest: wire.Digest,
	}
	canonical, err := snapshot.CanonicalJSON()
	if err != nil {
		return Snapshot{}, err
	}
	if !bytes.Equal(canonical, data) {
		return Snapshot{}, fail(CodeInvalidSchema, "snapshot JSON is not canonical")
	}
	return snapshot, nil
}
