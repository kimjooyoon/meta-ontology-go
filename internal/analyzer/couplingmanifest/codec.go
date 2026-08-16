package couplingmanifest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

// CanonicalJSON returns the strict canonical JSON representation of a
// digest-sealed manifest.
func (m Manifest) CanonicalJSON() ([]byte, error) {
	normalized, unsigned, err := canonicalUnsigned(m)
	if err != nil {
		return nil, err
	}
	if !validDigest(m.Digest) || digest(unsigned) != m.Digest {
		return nil, codecError(CodeInvalidManifest, "manifest digest does not match canonical content")
	}
	return json.Marshal(wireForManifest(normalized))
}

// StableHash returns the digest bound into the canonical manifest.
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
	Status               Status          `json:"status"`
	ObservationComplete  bool            `json:"observation_complete"`
	FullSuiteRequired    bool            `json:"full_suite_required"`
	BeforeSnapshotDigest string          `json:"before_snapshot_digest"`
	HeadSnapshotDigest   string          `json:"head_snapshot_digest"`
	RegistryDigest       string          `json:"registry_digest"`
	SourceMapDigest      string          `json:"source_map_digest"`
	ToolchainDigest      string          `json:"toolchain_digest"`
	ProfileDigest        string          `json:"profile_digest"`
	ReasonCode           ErrorCode       `json:"reason_code"`
	ResolvedSurfaceIDs   []string        `json:"resolved_surface_ids"`
	Entries              []ManifestEntry `json:"entries"`
	Counts               ComponentCounts `json:"counts"`
	Work                 Work            `json:"work"`
	Digest               string          `json:"digest"`
}

type unsignedManifestWire struct {
	Schema               string          `json:"schema"`
	Status               Status          `json:"status"`
	ObservationComplete  bool            `json:"observation_complete"`
	FullSuiteRequired    bool            `json:"full_suite_required"`
	BeforeSnapshotDigest string          `json:"before_snapshot_digest"`
	HeadSnapshotDigest   string          `json:"head_snapshot_digest"`
	RegistryDigest       string          `json:"registry_digest"`
	SourceMapDigest      string          `json:"source_map_digest"`
	ToolchainDigest      string          `json:"toolchain_digest"`
	ProfileDigest        string          `json:"profile_digest"`
	ReasonCode           ErrorCode       `json:"reason_code"`
	ResolvedSurfaceIDs   []string        `json:"resolved_surface_ids"`
	Entries              []ManifestEntry `json:"entries"`
	Counts               ComponentCounts `json:"counts"`
	Work                 Work            `json:"work"`
}

func wireForManifest(m Manifest) manifestWire {
	return manifestWire{Schema: m.Schema, Status: m.Status, ObservationComplete: m.ObservationComplete, FullSuiteRequired: m.FullSuiteRequired, BeforeSnapshotDigest: m.BeforeSnapshotDigest, HeadSnapshotDigest: m.HeadSnapshotDigest, RegistryDigest: m.RegistryDigest, SourceMapDigest: m.SourceMapDigest, ToolchainDigest: m.ToolchainDigest, ProfileDigest: m.ProfileDigest, ReasonCode: m.ReasonCode, ResolvedSurfaceIDs: m.ResolvedSurfaceIDs, Entries: m.Entries, Counts: m.Counts, Work: m.Work, Digest: m.Digest}
}

func manifestFromWire(wire manifestWire) Manifest {
	return Manifest{Schema: wire.Schema, Status: wire.Status, ObservationComplete: wire.ObservationComplete, FullSuiteRequired: wire.FullSuiteRequired, BeforeSnapshotDigest: wire.BeforeSnapshotDigest, HeadSnapshotDigest: wire.HeadSnapshotDigest, RegistryDigest: wire.RegistryDigest, SourceMapDigest: wire.SourceMapDigest, ToolchainDigest: wire.ToolchainDigest, ProfileDigest: wire.ProfileDigest, ReasonCode: wire.ReasonCode, ResolvedSurfaceIDs: wire.ResolvedSurfaceIDs, Entries: wire.Entries, Counts: wire.Counts, Work: wire.Work, Digest: wire.Digest}
}

func canonicalUnsigned(m Manifest) (Manifest, []byte, error) {
	normalized, err := normalizeManifest(m)
	if err != nil {
		return Manifest{}, nil, err
	}
	unsigned, err := json.Marshal(unsignedManifestWire{Schema: normalized.Schema, Status: normalized.Status, ObservationComplete: normalized.ObservationComplete, FullSuiteRequired: normalized.FullSuiteRequired, BeforeSnapshotDigest: normalized.BeforeSnapshotDigest, HeadSnapshotDigest: normalized.HeadSnapshotDigest, RegistryDigest: normalized.RegistryDigest, SourceMapDigest: normalized.SourceMapDigest, ToolchainDigest: normalized.ToolchainDigest, ProfileDigest: normalized.ProfileDigest, ReasonCode: normalized.ReasonCode, ResolvedSurfaceIDs: normalized.ResolvedSurfaceIDs, Entries: normalized.Entries, Counts: normalized.Counts, Work: normalized.Work})
	if err != nil {
		return Manifest{}, nil, codecError(CodeInvalidManifest, "encode unsigned manifest: %v", err)
	}
	return normalized, unsigned, nil
}

func normalizeManifest(m Manifest) (Manifest, error) {
	if m.Schema != SchemaV1 {
		return Manifest{}, codecError(CodeInvalidSchema, "schema %q is not %q", m.Schema, SchemaV1)
	}
	switch m.Status {
	case StatusComplete:
		if !m.ObservationComplete || m.FullSuiteRequired || m.ReasonCode != "" {
			return Manifest{}, codecError(CodeInvalidStatus, "COMPLETE manifest flags or reason are inconsistent")
		}
		for value, label := range map[string]string{m.BeforeSnapshotDigest: "before snapshot digest", m.HeadSnapshotDigest: "head snapshot digest", m.RegistryDigest: "registry digest", m.SourceMapDigest: "source-map digest", m.ToolchainDigest: "toolchain digest", m.ProfileDigest: "profile digest"} {
			if !validDigest(value) {
				return Manifest{}, codecError(CodeInvalidManifest, "%s is malformed", label)
			}
		}
	case StatusUnknown, StatusFailClosed:
		if m.ObservationComplete || !m.FullSuiteRequired || !validReason(m.ReasonCode) {
			return Manifest{}, codecError(CodeInvalidStatus, "non-complete manifest flags or reason are inconsistent")
		}
		for _, value := range []string{m.BeforeSnapshotDigest, m.HeadSnapshotDigest, m.RegistryDigest, m.SourceMapDigest, m.ToolchainDigest, m.ProfileDigest} {
			if value != "" && !validDigest(value) {
				return Manifest{}, codecError(CodeInvalidManifest, "non-complete manifest has a malformed digest")
			}
		}
	default:
		return Manifest{}, codecError(CodeInvalidStatus, "unknown manifest status %q", m.Status)
	}
	if m.ResolvedSurfaceIDs == nil || m.Entries == nil {
		return Manifest{}, codecError(CodeInvalidManifest, "resolved surface IDs and entries must be explicit arrays")
	}
	m.ResolvedSurfaceIDs = append([]string(nil), m.ResolvedSurfaceIDs...)
	sort.Strings(m.ResolvedSurfaceIDs)
	if !uniqueIDs(m.ResolvedSurfaceIDs) {
		return Manifest{}, codecError(CodeInvalidManifest, "resolved surface IDs are duplicated or malformed")
	}
	m.Entries = append([]ManifestEntry(nil), m.Entries...)
	sort.Slice(m.Entries, func(i, j int) bool { return m.Entries[i].SurfaceID < m.Entries[j].SurfaceID })
	if err := normalizeEntries(m.Entries); err != nil {
		return Manifest{}, err
	}
	if len(m.ResolvedSurfaceIDs) != len(m.Entries) {
		return Manifest{}, codecError(CodeInvalidManifest, "resolved surface IDs and entries are not the same complete set")
	}
	for i, entry := range m.Entries {
		if m.ResolvedSurfaceIDs[i] != entry.SurfaceID {
			return Manifest{}, codecError(CodeInvalidManifest, "resolved surface IDs and entries are not aligned")
		}
	}
	if m.Counts != countEntries(m.Entries) || m.Work.ComponentCount != len(m.Entries) || m.Work.WorkUnits != len(m.Entries) {
		return Manifest{}, codecError(CodeInvalidManifest, "component counts or work are inconsistent")
	}
	return m, nil
}

func countEntries(entries []ManifestEntry) ComponentCounts {
	counts := ComponentCounts{Registered: len(entries), Resolved: len(entries)}
	for _, entry := range entries {
		if entry.BeforePresent {
			counts.Before++
		}
		if entry.AfterPresent {
			counts.Head++
		}
	}
	return counts
}

func normalizeEntries(entries []ManifestEntry) error {
	seen := make(map[string]struct{}, len(entries))
	seenSymbols := make(map[string]struct{}, len(entries))
	seenOwners := make(map[string]struct{}, len(entries))
	seenMaps := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, err := normalizeID(entry.SurfaceID); err != nil {
			return codecError(CodeInvalidManifest, "entry has malformed surface ID")
		}
		if _, exists := seen[entry.SurfaceID]; exists {
			return codecError(CodeInvalidManifest, "entry surface IDs are duplicated")
		}
		seen[entry.SurfaceID] = struct{}{}
		for _, value := range []string{entry.CodeSymbolID, entry.SemanticOwnerID, entry.SourceMapID} {
			if _, err := normalizeID(value); err != nil {
				return codecError(CodeInvalidManifest, "entry %q has a malformed identity tuple", entry.SurfaceID)
			}
		}
		if _, exists := seenSymbols[entry.CodeSymbolID]; exists {
			return codecError(CodeInvalidManifest, "entry code symbol IDs are duplicated")
		}
		if _, exists := seenOwners[entry.SemanticOwnerID]; exists {
			return codecError(CodeInvalidManifest, "entry semantic owner IDs are duplicated")
		}
		if _, exists := seenMaps[entry.SourceMapID]; exists {
			return codecError(CodeInvalidManifest, "entry source-map IDs are duplicated")
		}
		seenSymbols[entry.CodeSymbolID] = struct{}{}
		seenOwners[entry.SemanticOwnerID] = struct{}{}
		seenMaps[entry.SourceMapID] = struct{}{}
		if !entry.BeforePresent && (entry.BeforeBindingDigest != "" || entry.BeforeSourceMapBindingDigest != "" || entry.BeforeBlobDigest != "") || !entry.AfterPresent && (entry.AfterBindingDigest != "" || entry.AfterSourceMapBindingDigest != "" || entry.AfterBlobDigest != "") {
			return codecError(CodeInvalidManifest, "entry %q has a digest for an absent side", entry.SurfaceID)
		}
		if entry.BeforePresent && (!validRawDigest(entry.BeforeBindingDigest) || !validRawDigest(entry.BeforeSourceMapBindingDigest) || !validDigest(entry.BeforeBlobDigest)) || entry.AfterPresent && (!validRawDigest(entry.AfterBindingDigest) || !validRawDigest(entry.AfterSourceMapBindingDigest) || !validDigest(entry.AfterBlobDigest)) {
			return codecError(CodeInvalidManifest, "entry %q has a malformed side digest", entry.SurfaceID)
		}
		if !entry.BeforePresent && !entry.AfterPresent {
			return codecError(CodeInvalidManifest, "entry %q is absent from both sides", entry.SurfaceID)
		}
	}
	return nil
}

func uniqueIDs(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, err := normalizeID(value); err != nil {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validReason(code ErrorCode) bool {
	switch code {
	case CodeMissingSnapshot, CodeMissingAuthority, CodeInvalidSnapshot, CodeAuthorityDrift, CodeUnknownChangedSurface, CodeDuplicateBinding, CodeConflictingBinding, CodeMalformedBinding, CodeCandidateBinding, CodeDerivedBinding, CodeInvalidStatus, CodeInvalidManifest, CodeInvalidSchema, CodeNonCanonicalJSON:
		return true
	default:
		return false
	}
}

func codecError(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Status: StatusFailClosed, FullSuiteRequired: true}
}

func sealResult(m Manifest, supplied ...error) (Manifest, error) {
	normalized, unsigned, err := canonicalUnsigned(m)
	if err != nil {
		return Manifest{}, err
	}
	normalized.Digest = digest(unsigned)
	if len(supplied) != 0 {
		return normalized, supplied[0]
	}
	return normalized, nil
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate object field %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return nil
	}
}
