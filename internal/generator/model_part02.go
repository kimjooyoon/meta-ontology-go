package generator

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
	NameSource     SourceSpan `json:"name_source"`
	ProfileID      string     `json:"profile_id,omitempty"`
	ProfileVersion int        `json:"profile_version,omitempty"`
	ProfileDigest  string     `json:"profile_digest,omitempty"`
}
