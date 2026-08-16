package couplingmanifest

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
	if err := validateManifestDigests(m); err != nil {
		return Manifest{}, err
	}
	if m.SourceMapDigest != "" {
		if _, err := rawDigest(m.SourceMapDigest); err != nil {
			return Manifest{}, codecError(CodeInvalidManifest, "source-map digest is malformed")
		}
	}
	if err := validateAdapterMetadata(m); err != nil {
		return Manifest{}, err
	}
	m.HeadSnapshotDigest = m.AfterSnapshotDigest
	return m, nil
}

func validateManifestDigests(m Manifest) error {
	digests := []struct {
		value string
		name  string
	}{
		{m.RegistryDigest, "registry digest"}, {m.ToolchainDigest, "toolchain digest"},
		{m.ProfileDigest, "profile digest"}, {m.BeforeSnapshotDigest, "before snapshot digest"},
		{m.AfterSnapshotDigest, "after snapshot digest"},
	}
	if m.Complete {
		for _, digest := range digests {
			if _, err := rawDigest(digest.value); err != nil {
				return codecError(CodeInvalidManifest, "%s is malformed", digest.name)
			}
		}
		if m.ZeroChange != isZeroChange(m.Entries) {
			return codecError(CodeInvalidManifest, "zero-change claim is inconsistent with entries")
		}
		return nil
	}
	if m.ZeroChange || len(m.Entries) != 0 {
		return codecError(CodeInvalidStatus, "incomplete manifest must have zero-change false and no entries")
	}
	for _, digest := range digests {
		if digest.value != "" {
			if _, err := rawDigest(digest.value); err != nil {
				return codecError(CodeInvalidManifest, "incomplete manifest has a malformed digest")
			}
		}
	}
	return nil
}

func validateAdapterMetadata(m Manifest) error {
	if !m.statsKnown {
		return nil
	}
	if m.Status == "" {
		return codecError(CodeInvalidStatus, "adapter status is missing")
	}
	if m.Complete && (m.Status != StatusComplete || m.FullSuiteRequired || m.ReasonCode != "") {
		return codecError(CodeInvalidStatus, "complete adapter status is inconsistent")
	}
	if !m.Complete && (m.Status == StatusComplete || !m.FullSuiteRequired || !validReason(m.ReasonCode)) {
		return codecError(CodeInvalidStatus, "incomplete adapter status is inconsistent")
	}
	if m.Counts != countEntries(m.Entries) || m.Work.ComponentCount != len(m.Entries) || m.Work.WorkUnits != len(m.Entries) {
		return codecError(CodeInvalidManifest, "component counts or work are inconsistent")
	}
	return nil
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
		if err := validateEntryIDs(entry); err != nil {
			return err
		}
		if err := validateUniqueEntryIDs(entry, seenSymbols, seenOwners); err != nil {
			return err
		}
		seenSymbols[entry.CodeSymbolID] = struct{}{}
		seenOwners[entry.SemanticOwnerID] = struct{}{}
		if err := validateEntryDigests(entry); err != nil {
			return err
		}
	}
	return nil
}

func validateEntryIDs(entry ManifestEntry) error {
	for _, value := range []struct {
		id   semantic.ID
		name string
	}{
		{entry.CodeSymbolID, "code symbol ID"}, {entry.SemanticOwnerID, "semantic owner ID"},
	} {
		if _, err := normalizeID(value.id); err != nil {
			return codecError(CodeInvalidManifest, "entry %q has malformed %s", entry.SurfaceID, value.name)
		}
	}
	return nil
}

func validateUniqueEntryIDs(entry ManifestEntry, symbols, owners map[semantic.ID]struct{}) error {
	if _, exists := symbols[entry.CodeSymbolID]; exists {
		return codecError(CodeInvalidManifest, "entry code symbol IDs are duplicated")
	}
	if _, exists := owners[entry.SemanticOwnerID]; exists {
		return codecError(CodeInvalidManifest, "entry semantic owner IDs are duplicated")
	}
	return nil
}

func validateEntryDigests(entry ManifestEntry) error {
	for _, value := range []struct {
		value string
		name  string
	}{
		{entry.BeforeBindingDigest, "before binding digest"}, {entry.AfterBindingDigest, "after binding digest"},
		{entry.BeforeBlobDigest, "before blob digest"}, {entry.AfterBlobDigest, "after blob digest"},
	} {
		if _, err := rawDigest(value.value); err != nil {
			return codecError(CodeInvalidManifest, "entry %q has a malformed %s", entry.SurfaceID, value.name)
		}
	}
	if entry.BeforeSourcePath == "" || entry.AfterSourcePath == "" {
		return codecError(CodeInvalidManifest, "entry %q has a missing source path", entry.SurfaceID)
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
