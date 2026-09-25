package analyzer

import (
	"sort"
)

// SourceFile is one Go source view supplied to AnalyzePackage.
type SourceFile struct {
	Filename    string
	PackagePath string
	Source      []byte
}

// DiagnosticCode identifies a deterministic analyzer diagnostic.
type DiagnosticCode string

const (
	DiagInvalidAnnotation     DiagnosticCode = "analyzer.invalid-annotation"
	DiagConflictingAnnotation DiagnosticCode = "analyzer.conflicting-annotation"
)

// Diagnostic describes a source-backed semantic-analysis problem. Invalid
// annotations never become registrations or semantic facts.
type Diagnostic struct {
	Code    DiagnosticCode
	Message string
	Span    Span
}

// String formats a diagnostic deterministically.
func (d Diagnostic) String() string {
	return d.Span.String() + ": " + string(d.Code) + ": " + d.Message
}

// Error implements error for convenient single-diagnostic reporting.
func (d Diagnostic) Error() string { return d.String() }

// Diagnostics is an ordered list of analyzer diagnostics.
type Diagnostics []Diagnostic

// SortBySpan returns a detached, deterministic source order. Filename is part
// of the ordering so package analysis does not depend on input file order.
func (d Diagnostics) SortBySpan() Diagnostics {
	result := append(Diagnostics(nil), d...)
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Span.Filename != right.Span.Filename {
			return left.Span.Filename < right.Span.Filename
		}
		if left.Span.Start.Offset != right.Span.Start.Offset {
			return left.Span.Start.Offset < right.Span.Start.Offset
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		return left.Message < right.Message
	})
	return result
}

// HasErrors reports whether at least one diagnostic was emitted.
func (d Diagnostics) HasErrors() bool { return len(d) > 0 }
