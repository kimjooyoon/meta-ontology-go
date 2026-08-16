// Package couplingmanifest builds detector-owned coupling.ChangeManifest
// values from source-backed analyzer observations. Detector authority stays in
// internal/detection/coupling; this package owns only source-map adaptation and
// construction metadata.
package couplingmanifest

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const absentDigest = "0000000000000000000000000000000000000000000000000000000000000000"

// Detector-owned aliases make the public adapter surface use the exact
// detector contracts instead of declaring look-alike authority or manifest
// structs.
const (
	SchemaV1          = detector.ManifestSchemaV1
	ManifestSchemaV1  = detector.ManifestSchemaV1
	FormatVersion     = detector.ManifestSchemaV1
	InputSchemaV1     = detector.InputSchemaV1
	ResultSchemaV1    = detector.ResultSchemaV1
	AuthoritySchemaV1 = detector.AuthorityContextSchemaV1
	RegistrySchemaV1  = detector.RegistrySchemaV1
	ConfigSchemaV1    = detector.ConfigSchemaV1
	ResourceSchemaV1  = detector.ResourceSchemaV1
	ReceiptSchemaV1   = detector.ReceiptSchemaV1
	BaselineSchemaV1  = detector.BaselineSchemaV1
)

type AuthorityContext = detector.AuthorityContext
type Registry = detector.Registry
type Surface = detector.Surface
type SourceMapBinding = detector.SourceMapBinding
type ManifestEntry = detector.ManifestEntry
type Manifest = detector.ChangeManifest
type ChangeManifest = detector.ChangeManifest
type DetectorResult = detector.Result
type Result = detector.Result

// ConstructionStatus is adapter metadata and never participates in detector
// result or manifest identity.
type ConstructionStatus string

const (
	ConstructionComplete   ConstructionStatus = "COMPLETE"
	ConstructionUnknown    ConstructionStatus = "UNKNOWN"
	ConstructionFailClosed ConstructionStatus = "FAIL_CLOSED"
)

// ConstructionCode identifies source-map construction failures. Detector
// reason codes remain detector-owned and are returned through detector.Result.
type ConstructionCode string

const (
	CodeMissingSnapshot       ConstructionCode = "couplingmanifest.missing-snapshot"
	CodeMissingAuthority      ConstructionCode = "couplingmanifest.missing-authority"
	CodeInvalidSnapshot       ConstructionCode = "couplingmanifest.invalid-snapshot"
	CodeAuthorityDrift        ConstructionCode = "couplingmanifest.authority-drift"
	CodeUnknownChangedSurface ConstructionCode = "couplingmanifest.unknown-changed-surface"
	CodeDuplicateBinding      ConstructionCode = "couplingmanifest.duplicate-binding"
	CodeConflictingBinding    ConstructionCode = "couplingmanifest.conflicting-binding"
	CodeMalformedBinding      ConstructionCode = "couplingmanifest.malformed-binding"
	CodeCandidateBinding      ConstructionCode = "couplingmanifest.candidate-binding"
	CodeDerivedBinding        ConstructionCode = "couplingmanifest.derived-binding"
)

// ConstructionError reports an adapter observation problem without pretending
// to be a detector decision.
type ConstructionError struct {
	Code              ConstructionCode
	Message           string
	Status            ConstructionStatus
	FullSuiteRequired bool
}

func (e *ConstructionError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// SourceMapObservation is adapter metadata linking one analyzer observation
// to the exact detector registry identity tuple. It does not replace a
// detector Surface or change detector authority.
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

type SourceMapRecord = SourceMapObservation
type Observation = SourceMapObservation

// SourceMapContext contains only source-map observations and its independent
// adapter digest. Registry, policy, snapshot, receipt, proof, and path
// authority are supplied by detector.AuthorityContext or detector.Input.
type SourceMapContext struct {
	Digest            string
	Before            []SourceMapObservation
	Head              []SourceMapObservation
	CandidateBindings []SourceMapObservation
	DerivedBindings   []SourceMapObservation
}

type Input struct {
	Before    *selectiveci.Snapshot
	Head      *selectiveci.Snapshot
	Authority detector.AuthorityContext
	SourceMap SourceMapContext
}

type ManifestInput = Input

// Metadata is construction-only information. It is deliberately separate
// from detector.ChangeManifest and detector.Result.
type Metadata struct {
	Status             ConstructionStatus
	Reason             ConstructionCode
	SourceMapDigest    string
	ResolvedSurfaceIDs []semantic.ID
	Counts             ComponentCounts
	Work               Work
}

type ComponentCounts struct {
	Registered int
	Before     int
	Head       int
	Resolved   int
}

type Work struct {
	ComponentCount int
	WorkUnits      int
}
