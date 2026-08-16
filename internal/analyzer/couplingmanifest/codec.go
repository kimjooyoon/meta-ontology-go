package couplingmanifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// CanonicalJSON returns the strict detector ChangeManifest JSON representation.
// Source paths are serialized for detector validation but are excluded from the
// digest identity, as required by the detector contract.
func (m Manifest) CanonicalJSON() ([]byte, error) {
	normalized, _, err := canonicalUnsigned(m)
	if err != nil {
		return nil, err
	}
	if _, err := rawDigest(m.Digest); err != nil || stableDigest(manifestCanonical(normalized)) != m.Digest {
		return nil, codecError(CodeInvalidManifest, "manifest digest does not match canonical content")
	}
	return json.Marshal(wireForManifest(normalized))
}

// StableHash returns the raw SHA-256 digest bound into the manifest.
func (m Manifest) StableHash() string { return m.Digest }

// Canonical returns canonical JSON or an empty string when the manifest is
// invalid.
func (m Manifest) Canonical() string {
	data, err := m.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(data)
}

// Validate checks the strict manifest and its digest without external state.
func (m Manifest) Validate() error {
	_, err := m.CanonicalJSON()
	return err
}

// MarshalJSON makes encoding/json use the strict canonical form.
func (m Manifest) MarshalJSON() ([]byte, error) { return m.CanonicalJSON() }

// UnmarshalJSON accepts only one exact canonical manifest.
func (m *Manifest) UnmarshalJSON(data []byte) error {
	decoded, err := DecodeJSON(data)
	if err != nil {
		return err
	}
	*m = decoded
	return nil
}

// DecodeJSON decodes one strict, versioned, canonical JSON manifest.
func DecodeJSON(data []byte) (Manifest, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Manifest{}, codecError(CodeInvalidSchema, "%v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire *manifestWire
	if err := decoder.Decode(&wire); err != nil {
		return Manifest{}, codecError(CodeInvalidSchema, "decode manifest JSON: %v", err)
	}
	if wire == nil {
		return Manifest{}, codecError(CodeInvalidSchema, "manifest JSON must be an object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Manifest{}, codecError(CodeInvalidSchema, "manifest JSON has trailing values")
		}
		return Manifest{}, codecError(CodeInvalidSchema, "decode manifest JSON after object: %v", err)
	}
	manifest := manifestFromWire(*wire)
	canonical, err := manifest.CanonicalJSON()
	if err != nil {
		return Manifest{}, err
	}
	if !bytes.Equal(canonical, data) {
		return Manifest{}, codecError(CodeNonCanonicalJSON, "manifest JSON is not canonical")
	}
	return manifest, nil
}

// Decode is the concise spelling of DecodeJSON.
func Decode(data []byte) (Manifest, error) { return DecodeJSON(data) }

// DecodeManifestJSON is the descriptive spelling of DecodeJSON.
func DecodeManifestJSON(data []byte) (Manifest, error) { return DecodeJSON(data) }

// EncodeJSON returns the exact canonical JSON representation without a
// trailing newline.
func EncodeJSON(manifest Manifest) ([]byte, error) { return manifest.CanonicalJSON() }

// EncodeManifestJSON is the descriptive spelling of EncodeJSON.
func EncodeManifestJSON(manifest Manifest) ([]byte, error) { return EncodeJSON(manifest) }

type manifestWire struct {
	Schema               string          `json:"schema"`
	Complete             bool            `json:"complete"`
	ZeroChange           bool            `json:"zero_change"`
	RegistryDigest       string          `json:"registry_digest"`
	ToolchainDigest      string          `json:"toolchain_digest"`
	ProfileDigest        string          `json:"profile_digest"`
	BeforeSnapshotDigest string          `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string          `json:"after_snapshot_digest"`
	Entries              []ManifestEntry `json:"entries"`
	Digest               string          `json:"digest"`
}

type unsignedManifestWire struct {
	Schema               string          `json:"schema"`
	Complete             bool            `json:"complete"`
	ZeroChange           bool            `json:"zero_change"`
	RegistryDigest       string          `json:"registry_digest"`
	ToolchainDigest      string          `json:"toolchain_digest"`
	ProfileDigest        string          `json:"profile_digest"`
	BeforeSnapshotDigest string          `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string          `json:"after_snapshot_digest"`
	Entries              []ManifestEntry `json:"entries"`
}

func wireForManifest(m Manifest) manifestWire {
	return manifestWire{
		Schema: m.Schema, Complete: m.Complete, ZeroChange: m.ZeroChange,
		RegistryDigest: m.RegistryDigest, ToolchainDigest: m.ToolchainDigest,
		ProfileDigest: m.ProfileDigest, BeforeSnapshotDigest: m.BeforeSnapshotDigest,
		AfterSnapshotDigest: m.AfterSnapshotDigest, Entries: m.Entries, Digest: m.Digest,
	}
}

func manifestFromWire(wire manifestWire) Manifest {
	manifest := Manifest{
		Schema: wire.Schema, Complete: wire.Complete, ZeroChange: wire.ZeroChange,
		RegistryDigest: wire.RegistryDigest, ToolchainDigest: wire.ToolchainDigest,
		ProfileDigest: wire.ProfileDigest, BeforeSnapshotDigest: wire.BeforeSnapshotDigest,
		AfterSnapshotDigest: wire.AfterSnapshotDigest, Entries: wire.Entries,
		Digest: wire.Digest, HeadSnapshotDigest: wire.AfterSnapshotDigest,
		Status: StatusUnknown, FullSuiteRequired: !wire.Complete,
	}
	if wire.Complete {
		manifest.Status = StatusComplete
	}
	manifest.ResolvedSurfaceIDs = make([]semantic.ID, 0, len(wire.Entries))
	for _, entry := range wire.Entries {
		manifest.ResolvedSurfaceIDs = append(manifest.ResolvedSurfaceIDs, entry.SurfaceID)
	}
	manifest.Counts = countEntries(wire.Entries)
	manifest.Work = Work{ComponentCount: len(wire.Entries), WorkUnits: len(wire.Entries)}
	return manifest
}

func canonicalUnsigned(m Manifest) (Manifest, []byte, error) {
	normalized, err := normalizeManifest(m)
	if err != nil {
		return Manifest{}, nil, err
	}
	unsigned, err := json.Marshal(unsignedManifestWire{
		Schema: normalized.Schema, Complete: normalized.Complete, ZeroChange: normalized.ZeroChange,
		RegistryDigest: normalized.RegistryDigest, ToolchainDigest: normalized.ToolchainDigest,
		ProfileDigest: normalized.ProfileDigest, BeforeSnapshotDigest: normalized.BeforeSnapshotDigest,
		AfterSnapshotDigest: normalized.AfterSnapshotDigest, Entries: normalized.Entries,
	})
	if err != nil {
		return Manifest{}, nil, codecError(CodeInvalidManifest, "encode unsigned manifest: %v", err)
	}
	return normalized, unsigned, nil
}
