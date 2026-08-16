// Package couplingmanifest adapts source-backed analyzer snapshots into the
// ChangeManifest contract consumed by the coupling detector.
package couplingmanifest

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	// SchemaV1 is the detector-owned wire schema. The adapter deliberately does
	// not publish a second manifest schema.
	SchemaV1 = "gooo/code-semantic-coupling-manifest/v1"
	// ManifestSchemaV1 is the detector vocabulary spelling of SchemaV1.
	ManifestSchemaV1 = SchemaV1
	// FormatVersion is the descriptive spelling of SchemaV1.
	FormatVersion = SchemaV1
	// AuthoritySchemaV1 identifies the in-memory registry/source-map contract.
	AuthoritySchemaV1 = "gooo/coupling-authority/v1"
)

const absentDigest = "0000000000000000000000000000000000000000000000000000000000000000"

// Status describes the adapter construction state. It is construction
// metadata, not part of the detector ChangeManifest wire object.
type Status string

const (
	StatusComplete   Status = "COMPLETE"
	StatusUnknown    Status = "UNKNOWN"
	StatusFailClosed Status = "FAIL_CLOSED"

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

// SourceMapBinding is the exact detector registry binding shape. Presentation
// labels are retained for callers but never enter registry or manifest identity.
type SourceMapBinding struct {
	SourceMapID   semantic.ID `json:"source_map_id"`
	BindingDigest string      `json:"binding_digest"`
	PackageLabel  string      `json:"package_label,omitempty"`
	FileLabel     string      `json:"file_label,omitempty"`
	SourceSpan    string      `json:"source_span,omitempty"`
}

// Surface is the exact detector registry surface shape. The four stable IDs
// remain distinct; labels and names are not represented as identity.
type Surface struct {
	SurfaceID         semantic.ID      `json:"surface_id"`
	CodeSymbolID      semantic.ID      `json:"code_symbol_id"`
	SemanticOwnerID   semantic.ID      `json:"semantic_owner_id"`
	Binding           SourceMapBinding `json:"binding"`
	PresentationLabel string           `json:"presentation_label,omitempty"`
}

// SourceMapObservation binds one analyzer snapshot observation to a
// registered surface. BindingDigest is the analyzer/selectiveci binding
// digest; SourceMapBindingDigest is the detector registry/source-map digest.
// Path and role are source locators/validation data, not identity.
type SourceMapObservation struct {
	SurfaceID              semantic.ID
	CodeSymbolID           semantic.ID
	SemanticOwnerID        semantic.ID
	SourceMapID            semantic.ID
	Role                   semanticbinding.Role
	Path                   string
	BlobDigest             string
	BindingDigest          string
	SourceMapBindingDigest string
}

// SourceMapRecord and Observation are vocabulary aliases for side bindings.
type SourceMapRecord = SourceMapObservation
type Observation = SourceMapObservation

// RegistrySourceMap is the explicit authority binding for both snapshots.
// Before and Head must be non-nil, including for an explicit empty source set.
// Candidate and derived records are observations, never authority.
type RegistrySourceMap struct {
	Schema            string
	RegistryDigest    string
	SourceMapDigest   string
	ToolchainDigest   string
	ProfileDigest     string
	Inventory         []Surface
	Before            []SourceMapObservation
	Head              []SourceMapObservation
	CandidateBindings []SourceMapObservation
	DerivedBindings   []SourceMapObservation
}

type Authority = RegistrySourceMap
type Registry = RegistrySourceMap
type SourceMap = RegistrySourceMap

// Input is the pure adapter input. Nil snapshots mean that the corresponding
// authority observation is unavailable; they are not an explicit zero state.
type Input struct {
	Before    *selectiveci.Snapshot
	Head      *selectiveci.Snapshot
	Authority RegistrySourceMap
}

type ManifestInput = Input

// ManifestEntry is the exact detector ManifestEntry contract. The additional
// fields are adapter-only metadata and are omitted from the wire object.
type ManifestEntry struct {
	SurfaceID           semantic.ID `json:"surface_id"`
	CodeSymbolID        semantic.ID `json:"code_symbol_id"`
	SemanticOwnerID     semantic.ID `json:"semantic_owner_id"`
	BeforeBindingDigest string      `json:"before_binding_digest"`
	AfterBindingDigest  string      `json:"after_binding_digest"`
	BeforeBlobDigest    string      `json:"before_blob_digest"`
	AfterBlobDigest     string      `json:"after_blob_digest"`
	BeforeSourcePath    string      `json:"before_source_path"`
	AfterSourcePath     string      `json:"after_source_path"`

	SourceMapID                  semantic.ID `json:"-"`
	BeforePresent                bool        `json:"-"`
	AfterPresent                 bool        `json:"-"`
	BeforeSourceMapBindingDigest string      `json:"-"`
	AfterSourceMapBindingDigest  string      `json:"-"`
}

// Component is retained as a vocabulary alias for detector entries.
type Component = ManifestEntry

// ComponentCounts are deterministic adapter metadata. They are recomputed
// from resolved entries and never participate in authorization.
type ComponentCounts struct {
	Registered int `json:"registered"`
	Before     int `json:"before"`
	Head       int `json:"head"`
	Resolved   int `json:"resolved"`
}

// Work is deterministic adapter metadata only.
type Work struct {
	ComponentCount int `json:"component_count"`
	WorkUnits      int `json:"work_units"`
}

// Manifest is the detector ChangeManifest wire contract. Status, source-map
// digest, resolved IDs, counts, and work remain adapter metadata and are not
// serialized as extra detector fields.
type Manifest struct {
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

	Status             Status          `json:"-"`
	FullSuiteRequired  bool            `json:"-"`
	ReasonCode         ErrorCode       `json:"-"`
	SourceMapDigest    string          `json:"-"`
	ResolvedSurfaceIDs []semantic.ID   `json:"-"`
	Counts             ComponentCounts `json:"-"`
	Work               Work            `json:"-"`
	HeadSnapshotDigest string          `json:"-"` // compatibility alias for callers using “head”.

	statsKnown bool
}

// ChangeManifest is the detector-facing vocabulary spelling.
type ChangeManifest = Manifest
