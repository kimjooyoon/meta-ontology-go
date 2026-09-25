package impactgraph

const SchemaVersion = "gooo/impact-graph/v1"

// NodeKind is the closed vocabulary for graph vertices.
type NodeKind string

const (
	NodeKindSource          NodeKind = "SOURCE"
	NodeKindSemantic        NodeKind = "SEMANTIC"
	NodeKindGoSymbol        NodeKind = "GO_SYMBOL"
	NodeKindGoPackage       NodeKind = "GO_PACKAGE"
	NodeKindGeneratedRegion NodeKind = "GENERATED_REGION"
	NodeKindObligation      NodeKind = "OBLIGATION"
	NodeKindPressure        NodeKind = "PRESSURE"

	NodeSource          = NodeKindSource
	NodeSemantic        = NodeKindSemantic
	NodeGoSymbol        = NodeKindGoSymbol
	NodeGoPackage       = NodeKindGoPackage
	NodeGeneratedRegion = NodeKindGeneratedRegion
	NodeObligation      = NodeKindObligation
	NodePressure        = NodeKindPressure
)

// EdgeKind is the closed vocabulary for directed graph edges.
type EdgeKind string

const (
	EdgeKindDeclares      EdgeKind = "DECLARES"
	EdgeKindImplements    EdgeKind = "IMPLEMENTS"
	EdgeKindProjectsTo    EdgeKind = "PROJECTS_TO"
	EdgeKindImportAffects EdgeKind = "IMPORT_AFFECTS"
	EdgeKindAffects       EdgeKind = "AFFECTS"
	EdgeKindVerifiedBy    EdgeKind = "VERIFIED_BY"
	EdgeKindMeasuredBy    EdgeKind = "MEASURED_BY"

	EdgeDeclares      = EdgeKindDeclares
	EdgeImplements    = EdgeKindImplements
	EdgeProjectsTo    = EdgeKindProjectsTo
	EdgeImportAffects = EdgeKindImportAffects
	EdgeAffects       = EdgeKindAffects
	EdgeVerifiedBy    = EdgeKindVerifiedBy
	EdgeMeasuredBy    = EdgeKindMeasuredBy
)

// Node is an identity-bearing graph vertex. ID is opaque and is never matched
// by display name, filename, or natural-language similarity.
type Node struct {
	ID   string   `json:"id"`
	Kind NodeKind `json:"kind"`
}

// Edge is a directed relation from From to To.
//
// Source, Target, Subject, and Object are construction-only aliases. They are
// excluded from the wire format so JSON remains one strict schema.
type Edge struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Kind EdgeKind `json:"kind"`

	Source  string `json:"-"`
	Target  string `json:"-"`
	Subject string `json:"-"`
	Object  string `json:"-"`
}
