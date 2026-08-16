// Package impactgraph provides a closed, digest-bound graph for deterministic
// change-impact obligation evaluation.
package impactgraph

import "errors"

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

// Graph is the complete digest-bound impact graph.
type Graph struct {
	Version        string `json:"version"`
	SnapshotDigest string `json:"snapshot_digest"`
	RegistryDigest string `json:"registry_digest"`
	PolicyDigest   string `json:"policy_digest"`
	Nodes          []Node `json:"nodes"`
	Edges          []Edge `json:"edges"`
}

// Decision is the closed evaluation outcome vocabulary.
type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionFailClosed Decision = "FAIL_CLOSED"
	DecisionUnknown    Decision = "UNKNOWN"

	PASS        = DecisionPass
	FAIL_CLOSED = DecisionFailClosed
	UNKNOWN     = DecisionUnknown
)

// Failure codes are stable machine-readable reasons for non-PASS outcomes.
const (
	FailureCodeNone                      = ""
	FailureCodeInvalidRegistry           = "INVALID_REGISTRY"
	FailureCodeUnknownChangedNode        = "UNKNOWN_CHANGED_NODE"
	FailureCodeUnknownExecutedObligation = "UNKNOWN_EXECUTED_OBLIGATION"
	FailureCodeAmbiguousChangedInput     = "AMBIGUOUS_CHANGED_INPUT"
	FailureCodeAmbiguousExecutedInput    = "AMBIGUOUS_EXECUTED_INPUT"
	FailureCodeNoReachableObligations    = "NO_REACHABLE_OBLIGATIONS"
	FailureCodeMissedObligations         = "MISSED_OBLIGATIONS"
)

var (
	ErrInvalidGraph    = errors.New("invalid impact graph")
	ErrInvalidNode     = errors.New("invalid impact graph node")
	ErrInvalidEdge     = errors.New("invalid impact graph edge")
	ErrDuplicateNode   = errors.New("duplicate impact graph node")
	ErrDuplicateEdge   = errors.New("duplicate impact graph edge")
	ErrInvalidDocument = errors.New("invalid impact graph document")
)

// Evaluation is the exact set comparison between reachable obligations and
// the obligations reported as executed.
type Evaluation struct {
	Required          []string `json:"required"`
	ExecutedRequired  []string `json:"executed_required"`
	Missed            []string `json:"missed"`
	Extra             []string `json:"extra"`
	Numerator         int      `json:"numerator"`
	Denominator       int      `json:"denominator"`
	Decision          Decision `json:"decision"`
	FailureCode       string   `json:"failure_code"`
	FullSuiteRequired bool     `json:"full_suite_required"`
}

// Result is an alternate name for the evaluation boundary.
type Result = Evaluation
