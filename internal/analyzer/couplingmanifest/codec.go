package couplingmanifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

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

func normalizeManifest(m Manifest) (Manifest, error) {
	if m.Schema != SchemaV1 {
		return Manifest{}, codecError(CodeInvalidSchema, "schema %q is not %q", m.Schema, SchemaV1)
	}
	if m.AfterSnapshotDigest == "" && m.HeadSnapshotDigest != "" {
		m.AfterSnapshotDigest = m.HeadSnapshotDigest
	}
	if m.HeadSnapshotDigest != "" && m.AfterSnapshotDigest != m.HeadSnapshotDigest {
		return Manifest{}, codecError(CodeInvalidManifest, "after/head snapshot digest aliases disagree")
	}
	if m.Entries == nil {
		return Manifest{}, codecError(CodeInvalidManifest, "manifest entries must be an explicit array")
	}
	m.Entries = append([]ManifestEntry(nil), m.Entries...)
	sort.Slice(m.Entries, func(i, j int) bool { return m.Entries[i].SurfaceID < m.Entries[j].SurfaceID })
	if err := normalizeEntries(m.Entries); err != nil {
		return Manifest{}, err
	}
	if m.Complete {
		for _, value := range []struct {
			value string
			name  string
		}{
			{m.RegistryDigest, "registry digest"},
			{m.ToolchainDigest, "toolchain digest"},
			{m.ProfileDigest, "profile digest"},
			{m.BeforeSnapshotDigest, "before snapshot digest"},
			{m.AfterSnapshotDigest, "after snapshot digest"},
		} {
			if _, err := rawDigest(value.value); err != nil {
				return Manifest{}, codecError(CodeInvalidManifest, "%s is malformed", value.name)
			}
		}
		if m.ZeroChange != isZeroChange(m.Entries) {
			return Manifest{}, codecError(CodeInvalidManifest, "zero-change claim is inconsistent with entries")
		}
	} else {
		if m.ZeroChange || len(m.Entries) != 0 {
			return Manifest{}, codecError(CodeInvalidStatus, "incomplete manifest must have zero-change false and no entries")
		}
		for _, value := range []string{m.RegistryDigest, m.ToolchainDigest, m.ProfileDigest, m.BeforeSnapshotDigest, m.AfterSnapshotDigest} {
			if value != "" {
				if _, err := rawDigest(value); err != nil {
					return Manifest{}, codecError(CodeInvalidManifest, "incomplete manifest has a malformed digest")
				}
			}
		}
	}
	if m.SourceMapDigest != "" {
		if _, err := rawDigest(m.SourceMapDigest); err != nil {
			return Manifest{}, codecError(CodeInvalidManifest, "source-map digest is malformed")
		}
	}
	if m.statsKnown {
		if m.Status == "" {
			return Manifest{}, codecError(CodeInvalidStatus, "adapter status is missing")
		}
		if m.Complete && (m.Status != StatusComplete || m.FullSuiteRequired || m.ReasonCode != "") {
			return Manifest{}, codecError(CodeInvalidStatus, "complete adapter status is inconsistent")
		}
		if !m.Complete && (m.Status == StatusComplete || !m.FullSuiteRequired || !validReason(m.ReasonCode)) {
			return Manifest{}, codecError(CodeInvalidStatus, "incomplete adapter status is inconsistent")
		}
		if m.Counts != countEntries(m.Entries) || m.Work.ComponentCount != len(m.Entries) || m.Work.WorkUnits != len(m.Entries) {
			return Manifest{}, codecError(CodeInvalidManifest, "component counts or work are inconsistent")
		}
	}
	m.HeadSnapshotDigest = m.AfterSnapshotDigest
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
	seen := make(map[semantic.ID]struct{}, len(entries))
	seenSymbols := make(map[semantic.ID]struct{}, len(entries))
	seenOwners := make(map[semantic.ID]struct{}, len(entries))
	for _, entry := range entries {
		if _, err := normalizeID(entry.SurfaceID); err != nil {
			return codecError(CodeInvalidManifest, "entry has malformed surface ID")
		}
		if _, exists := seen[entry.SurfaceID]; exists {
			return codecError(CodeInvalidManifest, "entry surface IDs are duplicated")
		}
		seen[entry.SurfaceID] = struct{}{}
		for _, value := range []struct {
			id   semantic.ID
			name string
		}{
			{entry.CodeSymbolID, "code symbol ID"},
			{entry.SemanticOwnerID, "semantic owner ID"},
		} {
			if _, err := normalizeID(value.id); err != nil {
				return codecError(CodeInvalidManifest, "entry %q has malformed %s", entry.SurfaceID, value.name)
			}
		}
		if _, exists := seenSymbols[entry.CodeSymbolID]; exists {
			return codecError(CodeInvalidManifest, "entry code symbol IDs are duplicated")
		}
		if _, exists := seenOwners[entry.SemanticOwnerID]; exists {
			return codecError(CodeInvalidManifest, "entry semantic owner IDs are duplicated")
		}
		seenSymbols[entry.CodeSymbolID] = struct{}{}
		seenOwners[entry.SemanticOwnerID] = struct{}{}
		for _, value := range []struct {
			value string
			name  string
		}{
			{entry.BeforeBindingDigest, "before binding digest"},
			{entry.AfterBindingDigest, "after binding digest"},
			{entry.BeforeBlobDigest, "before blob digest"},
			{entry.AfterBlobDigest, "after blob digest"},
		} {
			if _, err := rawDigest(value.value); err != nil {
				return codecError(CodeInvalidManifest, "entry %q has a malformed %s", entry.SurfaceID, value.name)
			}
		}
		if entry.BeforeSourcePath == "" || entry.AfterSourcePath == "" {
			return codecError(CodeInvalidManifest, "entry %q has a missing source path", entry.SurfaceID)
		}
	}
	return nil
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
	normalized, _, err := canonicalUnsigned(m)
	if err != nil {
		return Manifest{}, err
	}
	normalized.Digest = stableDigest(manifestCanonical(normalized))
	if len(supplied) != 0 {
		return normalized, supplied[0]
	}
	return normalized, nil
}

func field(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}

func manifestCanonical(manifest Manifest) string {
	entries := append([]ManifestEntry(nil), manifest.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	var builder strings.Builder
	field(&builder, SchemaV1)
	field(&builder, strconv.FormatBool(manifest.Complete))
	field(&builder, strconv.FormatBool(manifest.ZeroChange))
	field(&builder, manifest.RegistryDigest)
	field(&builder, manifest.ToolchainDigest)
	field(&builder, manifest.ProfileDigest)
	field(&builder, manifest.BeforeSnapshotDigest)
	field(&builder, manifest.AfterSnapshotDigest)
	for _, entry := range entries {
		field(&builder, entry.SurfaceID.String())
		field(&builder, entry.CodeSymbolID.String())
		field(&builder, entry.SemanticOwnerID.String())
		field(&builder, entry.BeforeBindingDigest)
		field(&builder, entry.AfterBindingDigest)
		field(&builder, entry.BeforeBlobDigest)
		field(&builder, entry.AfterBlobDigest)
	}
	return builder.String()
}

func stableDigest(value string) string { return semantic.StableHashString(value) }

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
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}
