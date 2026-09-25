package query

import (
	"errors"
)

const (
	// QueryEnvelopeSchema identifies the machine-readable request/response
	// contract. It is independent from the graph projection schema.
	QueryEnvelopeSchema = "gooo-query/v1"
	MaxEnvelopeDepth    = 64
	MaxEnvelopeLimit    = 1000
)

var (
	ErrInvalidEnvelope      = errors.New("invalid query envelope")
	ErrUnsupportedOperation = errors.New("unsupported query operation")
	ErrUnsupportedLayer     = errors.New("unsupported query layer")
	ErrUnsupportedDirection = errors.New("unsupported query direction")
	ErrAmbiguousTraversal   = errors.New("ambiguous query traversal")
	ErrInvalidEnvelopeLimit = errors.New("invalid query envelope limit")
	ErrInvalidEnvelopeDepth = errors.New("invalid query envelope depth")
)

// Operation selects the read operation represented by an envelope.
type Operation string

const (
	OperationExact     Operation = "exact"
	OperationTraversal Operation = "traverse"
	OperationDerived   Operation = "derived"
)

// Layer is the explicit fact universe selected by an envelope.
type Layer string

const (
	LayerAll           Layer = "all"
	LayerDeterministic Layer = "deterministic"
	LayerCandidate     Layer = "candidate"
)

// Request is a versioned, read-only query envelope. Target is required for an
// exact query and forbidden for traversal or derived rules. Empty Relation
// means all supported relations for traversal only; Rule is required by the
// derived operation and forbidden by exact/traversal operations.
type Request struct {
	Schema    string        `json:"schema"`
	Operation Operation     `json:"operation"`
	Root      ID            `json:"root"`
	Target    ID            `json:"target,omitempty"`
	Relation  Relation      `json:"relation,omitempty"`
	Rule      DerivedRuleID `json:"rule,omitempty"`
	Layer     Layer         `json:"layer"`
	Direction string        `json:"direction,omitempty"`
	MaxDepth  int           `json:"max_depth"`
	Limit     int           `json:"limit"`
}

// QueryResult keeps fact layers and traversal paths separate in a response.
type QueryResult struct {
	DeterministicMatches []Fact        `json:"deterministic_matches,omitempty"`
	CandidateMatches     []Fact        `json:"candidate_matches,omitempty"`
	DeterministicPaths   []Path        `json:"deterministic_paths,omitempty"`
	CandidatePaths       []Path        `json:"candidate_paths,omitempty"`
	DerivedDeterministic []DerivedFact `json:"derived_deterministic,omitempty"`
	DerivedCandidates    []DerivedFact `json:"derived_candidates,omitempty"`
}
