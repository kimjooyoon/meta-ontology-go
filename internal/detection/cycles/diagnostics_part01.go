package cycles

import (
	"fmt"
	"strings"
)

// Code identifies one graph invariant failure.
type Code string

const (
	CycleDetected            Code = "cycle"
	IllegalRelationDirection Code = "illegal-relation-direction"
	UnresolvedStableID       Code = "unresolved-stable-id"
	NamespaceCollision       Code = "namespace-collision"
	InvalidStableID          Code = "invalid-stable-id"

	CodeCycle                    = CycleDetected
	CodeIllegalRelationDirection = IllegalRelationDirection
	CodeUnresolvedStableID       = UnresolvedStableID
	CodeNamespaceCollision       = NamespaceCollision
	CodeInvalidStableID          = InvalidStableID
)

// Diagnostic is one deterministic graph validation result. Subject, Object,
// and Predicate identify an offending edge when applicable. Cycle contains a
// closed path for cycle diagnostics; other diagnostics leave it empty.
type Diagnostic struct {
	Code      Code
	Message   string
	NodeID    ID
	Namespace string
	Name      string
	Subject   ID
	Predicate Relation
	Object    ID
	Cycle     []ID
	Span      Span
}

// Issue is a descriptive alias for Diagnostic.
type Issue = Diagnostic

// Diagnostics is an ordered collection of graph diagnostics and implements
// error for convenient use by Check.
type Diagnostics []Diagnostic

// Error returns stable, line-oriented diagnostics. An empty collection is an
// empty error string, matching the usual nil-error convention for callers
// that inspect the result directly.
func (d Diagnostics) Error() string {
	if len(d) == 0 {
		return ""
	}
	lines := make([]string, len(d))
	for i, diagnostic := range d {
		lines[i] = diagnostic.Error()
	}
	return strings.Join(lines, "\n")
}

// Error formats one diagnostic without depending on process-local map order.
func (d Diagnostic) Error() string {
	if d.Message == "" {
		return string(d.Code)
	}
	if d.Code == "" {
		return d.Message
	}
	return fmt.Sprintf("%s: %s", d.Code, d.Message)
}
