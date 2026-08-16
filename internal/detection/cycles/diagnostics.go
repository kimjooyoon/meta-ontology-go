package cycles

import (
	"fmt"
	"sort"
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

// Has reports whether at least one diagnostic has code.
func (d Diagnostics) Has(code Code) bool {
	for _, diagnostic := range d {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

// Codes returns the distinct diagnostic codes in lexical order.
func (d Diagnostics) Codes() []Code {
	seen := make(map[Code]struct{}, len(d))
	for _, diagnostic := range d {
		seen[diagnostic.Code] = struct{}{}
	}
	result := make([]Code, 0, len(seen))
	for code := range seen {
		result = append(result, code)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func sortDiagnostics(diagnostics Diagnostics) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		left, right := diagnostics[i], diagnostics[j]
		leftKey := diagnosticKey(left)
		rightKey := diagnosticKey(right)
		return leftKey < rightKey
	})
}

func diagnosticKey(diagnostic Diagnostic) string {
	return strings.Join([]string{
		string(diagnostic.Code), diagnostic.Namespace, diagnostic.Name,
		diagnostic.NodeID, diagnostic.Subject, string(diagnostic.Predicate),
		diagnostic.Object, strings.Join(diagnostic.Cycle, "\x00"),
		diagnostic.Message, diagnostic.Span.File,
		fmt.Sprintf("%09d:%09d", diagnostic.Span.Start.Line, diagnostic.Span.Start.Column),
	}, "\x00")
}
