// Package couplingmanifest adapts source-backed analyzer snapshots into the
// identity-only change manifest consumed by the coupling detector.
package couplingmanifest

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
)

const (
	// SchemaV1 is the lossless, adapter-owned wire schema for Manifest.
	SchemaV1 = "gooo/coupling-change-manifest-adapter/v1"
	// FormatVersion is the descriptive spelling of SchemaV1.
	FormatVersion = SchemaV1
	// AuthoritySchemaV1 identifies the in-memory registry/source-map contract.
	AuthoritySchemaV1 = "gooo/coupling-authority/v1"
)

// Status is the closed construction state of a coupling manifest.
type Status string

const (
	StatusComplete   Status = "COMPLETE"
	StatusUnknown    Status = "UNKNOWN"
	StatusFailClosed Status = "FAIL_CLOSED"

	// Complete, Unknown, and FailClosed are concise compatibility spellings.
	Complete   = StatusComplete
	Unknown    = StatusUnknown
	FailClosed = StatusFailClosed
)

// ErrorCode identifies a deterministic construction or codec failure.
type ErrorCode string

const (
	CodeMissingSnapshot       ErrorCode = "couplingmanifest.missing-snapshot"
	CodeMissingAuthority      ErrorCode = "couplingmanifest.missing-authority"
	CodeInvalidSnapshot       ErrorCode = "couplingmanifest.invalid-snapshot"
	CodeAuthorityDrift        ErrorCode = "couplingmanifest.authority-drift"
	CodeUnknownChangedSurface ErrorCode = "couplingmanifest.unknown-changed-surface"
	CodeDuplicateBinding      ErrorCode = "couplingmanifest.duplicate-binding"
	CodeConflictingBinding    ErrorCode = "couplingmanifest.conflicting-binding"
	CodeMalformedBinding      ErrorCode = "couplingmanifest.malformed-binding"
	CodeCandidateBinding      ErrorCode = "couplingmanifest.candidate-binding"
	CodeDerivedBinding        ErrorCode = "couplingmanifest.derived-binding"
	CodeInvalidStatus         ErrorCode = "couplingmanifest.invalid-status"
	CodeInvalidManifest       ErrorCode = "couplingmanifest.invalid-manifest"
	CodeInvalidSchema         ErrorCode = "couplingmanifest.invalid-schema"
	CodeNonCanonicalJSON      ErrorCode = "couplingmanifest.non-canonical-json"
)

// Error is a deterministic result classification. Unknown and fail-closed
// construction errors never carry a partial resolved set.
type Error struct {
	Code              ErrorCode
	Message           string
	Status            Status
	FullSuiteRequired bool
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// SourceMapBinding is one explicit source-map/registry record for one side of
// a snapshot. Path is checked as a source locator, but never enters manifest
// identity. Every identity tuple member must be a stable semantic identifier.
type SourceMapBinding struct {
	SurfaceID              string
	CodeSymbolID           string
	SemanticOwnerID        string
	SourceMapID            string
	Role                   semanticbinding.Role
	Path                   string
	BlobDigest             string
	BindingDigest          string
	SourceMapBindingDigest string
}

// Binding is a vocabulary alias for SourceMapBinding.
type Binding = SourceMapBinding

// Surface is one registered authority identity. The inventory is complete and
// is independent of which side of the snapshot currently contains a source.
type Surface struct {
	SurfaceID       string
	CodeSymbolID    string
	SemanticOwnerID string
	SourceMapID     string
}

// RegistrySourceMap is the explicit authority binding for both snapshots.
// Before and Head must be non-nil, including when one side is an explicit
// empty source set. Candidate and derived records are never authoritative.
type RegistrySourceMap struct {
	Schema            string
	RegistryDigest    string
	SourceMapDigest   string
	ToolchainDigest   string
	ProfileDigest     string
	Inventory         []Surface
	Before            []SourceMapBinding
	Head              []SourceMapBinding
	CandidateBindings []SourceMapBinding
	DerivedBindings   []SourceMapBinding
}

// Authority is a descriptive alias for RegistrySourceMap.
type Authority = RegistrySourceMap

// Registry and SourceMap are concise vocabulary aliases for the explicit
// combined authority binding.
type Registry = RegistrySourceMap
type SourceMap = RegistrySourceMap

// Input is the pure adapter input. Nil snapshots mean that the corresponding
// authority observation is unavailable; they are not an explicit zero state.
type Input struct {
	Before    *selectiveci.Snapshot
	Head      *selectiveci.Snapshot
	Authority RegistrySourceMap
}

// ManifestInput is a vocabulary alias for Input.
type ManifestInput = Input

// Component is one resolved identity-bearing mechanical change. It contains
// exact source/blob binding digests, but no paths, names, aliases, labels, or
// prose.
type Component struct {
	SurfaceID                    string `json:"surface_id"`
	CodeSymbolID                 string `json:"code_symbol_id"`
	SemanticOwnerID              string `json:"semantic_owner_id"`
	SourceMapID                  string `json:"source_map_id"`
	BeforePresent                bool   `json:"before_present"`
	AfterPresent                 bool   `json:"after_present"`
	BeforeBindingDigest          string `json:"before_binding_digest"`
	AfterBindingDigest           string `json:"after_binding_digest"`
	BeforeSourceMapBindingDigest string `json:"before_source_map_binding_digest"`
	AfterSourceMapBindingDigest  string `json:"after_source_map_binding_digest"`
	BeforeBlobDigest             string `json:"before_blob_digest"`
	AfterBlobDigest              string `json:"after_blob_digest"`
}

// ManifestEntry is the detector-facing name for Component.
type ManifestEntry = Component

// ComponentCounts are deterministic counts only; they are not an
// authorization or resource decision.
type ComponentCounts struct {
	Registered int `json:"registered"`
	Before     int `json:"before"`
	Head       int `json:"head"`
	Resolved   int `json:"resolved"`
}

// Work is the deterministic amount of coupling-detector work described by a
// manifest. No command, authorization, or write instruction is represented.
type Work struct {
	ComponentCount int `json:"component_count"`
	WorkUnits      int `json:"work_units"`
}

// Manifest is the strict canonical coupling-detector input. A COMPLETE
// manifest with an explicit full inventory is distinct from UNKNOWN, whose
// resolved surface IDs and entries are empty.
type Manifest struct {
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

// ChangeManifest is a vocabulary alias for Manifest.
type ChangeManifest = Manifest
