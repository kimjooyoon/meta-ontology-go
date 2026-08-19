package couplingmanifest

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
