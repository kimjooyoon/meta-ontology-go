// Package generator contains the dependency-free Go projection boundary.
package generator

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

// Field is an optional structural field of an Entity.
type Field struct {
	ID     string
	Name   string
	GoName string
	GoType string
	Source SourceSpan
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
	SemanticID string
	Kind       string
	Source     SourceSpan
	Generated  SourceRange
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

// GenerationMetadata describes reproducible projection inputs and trust.
type GenerationMetadata struct {
	SourceDigest     string
	SemanticIRDigest string
	SourceMapDigest  string
	Evidence         EvidenceStatus
	Toolchain        ToolchainIdentity
	Projection       ProjectionStatus
}

// EvidenceStatus records what this package can prove without external files.
type EvidenceStatus struct {
	Decision string
	Refs     []string
}

// ToolchainIdentity is deferred because the generator does not own a build
// or receipt store. It never fabricates an environment identity.
type ToolchainIdentity struct {
	Status string
	Value  string
}

// ProjectionStatus records the authoritative status of this generator result.
type ProjectionStatus struct {
	Decision string
	Refs     []string
}

// Generator renders semantic input into Go source.
type Generator struct {
	Options Options
}

// New returns a Generator configured with options.
func New(options Options) Generator {
	return Generator{Options: options}
}
