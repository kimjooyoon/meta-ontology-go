package impactgraph

import (
	"errors"
)

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
