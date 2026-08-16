// Package generator contains the dependency-free Go projection boundary.
package generator

import "encoding/json"

// SemanticIR is the local DTO consumed by the Go projection. IDs are stable
// declaration identity; names are presentation and may change.
type SemanticIR struct {
	Package    string
	Imports    []Import
	Entities   []Entity
	Activities []Activity
}

// Import describes a Go import required by a projected type.
type Import struct {
	Path string
	Name string
}

// Entity is a semantic entity projected to a Go struct type.
type Entity struct {
	ID     string
	Name   string
	GoName string
	Fields []Field
	Source SourceSpan
}

// Field is a semantic structural field of an Entity. ID, Parent, source
// spans, and declaration order remain authoritative; Go projection details
// are derived only after the bound EntityFields profile is accepted.
type Field struct {
	ID              string
	Parent          string
	Name            string
	Aliases         []string
	TypeRefID       string
	Presence        string
	Cardinality     string
	Origin          string
	GoName          string
	GoType          string
	Source          SourceSpan
	IDSpan          SourceSpan
	NameSpan        SourceSpan
	TypeRefSpan     SourceSpan
	PresenceSpan    SourceSpan
	CardinalitySpan SourceSpan
	NameSource      SourceSpan // compatibility alias; NameSpan is authoritative
}

// Activity is projected to a Go function. Port order is part of its boundary.
type Activity struct {
	ID      string
	Name    string
	GoName  string
	Inputs  []Port
	Outputs []Port
	Slots   []Slot
	Source  SourceSpan
}

// Port describes an Activity input or output.
type Port struct {
	ID       string
	Name     string
	GoName   string
	EntityID string
	GoType   string
	Source   SourceSpan
}

// Slot is a handwritten implementation region inside a generated boundary.
// ID must remain stable across declaration renames.
type Slot struct {
	ID      string
	Name    string
	Default string
	Source  SourceSpan
}

// SourceSpan identifies the authoritative source declaration.
type SourceSpan struct {
	URI   string
	Start Position
	End   Position
}

// Position is a one-based source position with a byte offset.
type Position struct {
	Offset int
	Line   int
	Column int
}

// SourceRange is a half-open range in generated or source text.
type SourceRange struct {
	Start Position
	End   Position
}

// SourceMap maps semantic identities and handwritten slots to generated code.
type SourceMap struct {
	Mappings []SourceMapping
}

// SourceMapping is one semantic-to-generated range mapping.
type SourceMapping struct {
	SemanticID     string
	Kind           string
	Ordinal        int
	Source         SourceSpan
	Generated      SourceRange
	ParentID       string     `json:"parent_id,omitempty"`
	TypeRefID      string     `json:"type_ref_id,omitempty"`
	Presence       string     `json:"presence,omitempty"`
	Cardinality    string     `json:"cardinality,omitempty"`
	NameSource     SourceSpan `json:"name_source,omitempty"`
	ProfileID      string     `json:"profile_id,omitempty"`
	ProfileVersion int        `json:"profile_version,omitempty"`
	ProfileDigest  string     `json:"profile_digest,omitempty"`
}

// MarshalJSON keeps the established fieldless source-map wire form stable
// while publishing the additional field metadata only for field mappings.
func (mapping SourceMapping) MarshalJSON() ([]byte, error) {
	type sourceMappingWire struct {
		SemanticID     string      `json:"SemanticID"`
		Kind           string      `json:"Kind"`
		Ordinal        int         `json:"Ordinal"`
		Source         SourceSpan  `json:"Source"`
		Generated      SourceRange `json:"Generated"`
		ParentID       string      `json:"parent_id,omitempty"`
		TypeRefID      string      `json:"type_ref_id,omitempty"`
		Presence       string      `json:"presence,omitempty"`
		Cardinality    string      `json:"cardinality,omitempty"`
		NameSource     *SourceSpan `json:"name_source,omitempty"`
		ProfileID      string      `json:"profile_id,omitempty"`
		ProfileVersion int         `json:"profile_version,omitempty"`
		ProfileDigest  string      `json:"profile_digest,omitempty"`
	}
	var nameSource *SourceSpan
	if !sourceSpanIsZero(mapping.NameSource) {
		span := mapping.NameSource
		nameSource = &span
	}
	return json.Marshal(sourceMappingWire{
		SemanticID: mapping.SemanticID, Kind: mapping.Kind, Ordinal: mapping.Ordinal,
		Source: mapping.Source, Generated: mapping.Generated, ParentID: mapping.ParentID,
		TypeRefID: mapping.TypeRefID, Presence: mapping.Presence, Cardinality: mapping.Cardinality,
		NameSource: nameSource, ProfileID: mapping.ProfileID, ProfileVersion: mapping.ProfileVersion,
		ProfileDigest: mapping.ProfileDigest,
	})
}

// Lookup returns mappings for an identity in generated order.
func (m SourceMap) Lookup(semanticID string) []SourceMapping {
	var result []SourceMapping
	for _, mapping := range m.Mappings {
		if mapping.SemanticID == semanticID {
			result = append(result, mapping)
		}
	}
	return result
}

// Options controls projection details.
type Options struct {
	// Header is emitted only when generating a new file.
	Header string
	// PackageName overrides the package name supplied by an adapter.
	PackageName string
}

// Result is generated Go source together with its semantic source map.
type Result struct {
	Source    []byte
	SourceMap SourceMap
}

// MetadataResult is a typed projection result with deterministic identity
// digests. Unavailable external provenance is represented explicitly.
type MetadataResult struct {
	Result
	Metadata GenerationMetadata
}

// ProjectionMetadataV1 is the versioned wire surface for generator output.
type ProjectionMetadataV1 struct {
	Schema     string             `json:"schema"`
	Source     []byte             `json:"source"`
	SemanticIR SemanticIR         `json:"semantic_ir"`
	SourceMap  SourceMap          `json:"source_map"`
	Metadata   GenerationMetadata `json:"metadata"`
}

// ProjectionBinding supplies independently computed identities for opt-in
// fail-closed projection verification. Empty external fields stay deferred.
type ProjectionBinding struct {
	Schema           string            `json:"schema"`
	SourceDigest     string            `json:"source_digest"`
	SemanticIRDigest string            `json:"semantic_ir_digest"`
	SourceMapDigest  string            `json:"source_map_digest"`
	EvidenceDigest   string            `json:"evidence_digest,omitempty"`
	ProvenanceDigest string            `json:"provenance_digest,omitempty"`
	Toolchain        ToolchainIdentity `json:"toolchain"`
}

// GenerationMetadata describes reproducible projection inputs and trust.
type GenerationMetadata struct {
	SourceDigest     string                `json:"source_digest"`
	SemanticIRDigest string                `json:"semantic_ir_digest"`
	SourceMapDigest  string                `json:"source_map_digest"`
	Source           BindingStatus         `json:"source"`
	SemanticIR       BindingStatus         `json:"semantic_ir"`
	Provenance       BindingStatus         `json:"provenance"`
	Evidence         EvidenceStatus        `json:"evidence"`
	Toolchain        ToolchainIdentity     `json:"toolchain"`
	Projection       ProjectionStatus      `json:"projection"`
	Authority        AuthorityLabels       `json:"authority"`
	EntityFields     *EntityFieldsMetadata `json:"entity_fields,omitempty"`
}

// EntityFieldsMetadata records the exact profile bound to a supported field
// projection. It is nil for the fieldless/deferred production path so the
// established fieldless metadata wire representation remains unchanged.
type EntityFieldsMetadata struct {
	State   string                      `json:"state"`
	Profile EntityFieldsProfileMetadata `json:"profile"`
}

type EntityFieldsProfileMetadata struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Digest  string `json:"digest"`
}

// BindingStatus identifies whether a value is available and authoritative.
type BindingStatus struct {
	Status    string `json:"status"`
	Authority string `json:"authority"`
}

// AuthorityLabels names the verifier and projection boundaries explicitly.
type AuthorityLabels struct {
	Projection string `json:"projection"`
	Verifier   string `json:"verifier"`
	Provenance string `json:"provenance"`
}

// EvidenceStatus records what this package can prove without external files.
type EvidenceStatus struct {
	Decision string   `json:"decision"`
	Refs     []string `json:"refs"`
}

// ToolchainIdentity is deferred because the generator does not own a build
// or receipt store. It never fabricates an environment identity.
type ToolchainIdentity struct {
	Status string `json:"status"`
	Value  string `json:"value"`
}

// ProjectionStatus records the authoritative status of this generator result.
type ProjectionStatus struct {
	Decision string   `json:"decision"`
	Refs     []string `json:"refs"`
}

// Generator renders semantic input into Go source.
type Generator struct {
	Options Options
}

// New returns a Generator configured with options.
func New(options Options) Generator {
	return Generator{Options: options}
}
