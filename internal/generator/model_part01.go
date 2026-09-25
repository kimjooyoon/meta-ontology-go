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
