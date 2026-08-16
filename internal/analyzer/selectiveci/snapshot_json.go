package selectiveci

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

func decodeSnapshotWire(data []byte) (snapshotWire, error) {
	var wire snapshotWire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return snapshotWire{}, fail(CodeInvalidSchema, "decode snapshot JSON: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return snapshotWire{}, fail(CodeInvalidSchema, "snapshot JSON has trailing values")
		}
		return snapshotWire{}, fail(CodeInvalidSchema, "decode snapshot JSON after object: %v", err)
	}
	return wire, nil
}

type snapshotWire struct {
	Schema            string   `json:"schema"`
	Status            Status   `json:"status"`
	FullSuiteFallback bool     `json:"full_suite_fallback"`
	SourceMapDigest   string   `json:"source_map_digest"`
	RegistryDigest    string   `json:"registry_digest"`
	RegisteredIDs     []string `json:"registered_ids"`
	Sources           []Source `json:"sources"`
	Digest            string   `json:"digest"`
}

func wireForSnapshot(s Snapshot) snapshotWire {
	return snapshotWire{
		Schema: s.Schema, Status: s.Status, FullSuiteFallback: s.FullSuiteFallback,
		SourceMapDigest: s.SourceMapDigest, RegistryDigest: s.RegistryDigest,
		RegisteredIDs: s.RegisteredIDs, Sources: s.Sources, Digest: s.Digest,
	}
}

func (s Snapshot) unsignedJSON() ([]byte, error) {
	type unsignedWire struct {
		Schema            string   `json:"schema"`
		Status            Status   `json:"status"`
		FullSuiteFallback bool     `json:"full_suite_fallback"`
		SourceMapDigest   string   `json:"source_map_digest"`
		RegistryDigest    string   `json:"registry_digest"`
		RegisteredIDs     []string `json:"registered_ids"`
		Sources           []Source `json:"sources"`
	}
	return json.Marshal(unsignedWire{
		Schema: s.Schema, Status: s.Status, FullSuiteFallback: s.FullSuiteFallback,
		SourceMapDigest: s.SourceMapDigest, RegistryDigest: s.RegistryDigest,
		RegisteredIDs: s.RegisteredIDs, Sources: s.Sources,
	})
}

func normalizeSnapshot(s Snapshot) (Snapshot, error) {
	if s.Schema != SchemaV1 {
		return Snapshot{}, fail(CodeInvalidSchema, "schema %q is not %q", s.Schema, SchemaV1)
	}
	if s.Status != StatusBound || s.FullSuiteFallback || s.Digest == "" {
		return Snapshot{}, fail(CodeInvalidStatus, "snapshot must be BOUND with a digest")
	}
	if s.RegisteredIDs == nil || s.Sources == nil {
		return Snapshot{}, fail(CodeInvalidSchema, "registered IDs and sources must be explicit arrays")
	}
	registered, err := normalizeRegisteredIDs(s.RegisteredIDs)
	if err != nil {
		return Snapshot{}, err
	}
	sourceMapDigest, err := normalizeDigest(s.SourceMapDigest, "source-map digest")
	if err != nil {
		return Snapshot{}, err
	}
	registryDigest, err := normalizeDigest(s.RegistryDigest, "registry digest")
	if err != nil {
		return Snapshot{}, err
	}
	sources, err := normalizeManifestSources(s.Sources, registered)
	if err != nil {
		return Snapshot{}, err
	}
	if !validDigest(s.Digest) {
		return Snapshot{}, fail(CodeMalformedDigest, "snapshot digest is malformed")
	}
	s.RegisteredIDs, s.SourceMapDigest = sortedIDs(registered), sourceMapDigest
	s.RegistryDigest, s.Sources = registryDigest, sources
	unsigned, err := s.unsignedJSON()
	if err != nil {
		return Snapshot{}, err
	}
	if digest(unsigned) != s.Digest {
		return Snapshot{}, fail(CodeStaleSnapshot, "snapshot digest does not match canonical content")
	}
	return s, nil
}
