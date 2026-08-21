package analyzer

import (
	"fmt"
	"strings"
)

// Error returns all diagnostics as one deterministic error value.
func (d Diagnostics) Error() error {
	if !d.HasErrors() {
		return nil
	}
	lines := make([]string, 0, len(d))
	for _, diagnostic := range d.SortBySpan() {
		lines = append(lines, diagnostic.String())
	}
	return diagnosticError(strings.Join(lines, "\n"))
}

type diagnosticError string

func (e diagnosticError) Error() string { return string(e) }

// ObservationOrigin distinguishes contract-shaped signature facts from
// implementation observations in a generated Go projection.
type ObservationOrigin string

const (
	OriginSignature      ObservationOrigin = "signature"
	OriginImplementation ObservationOrigin = "implementation"
)

// Position is a zero-based byte offset plus one-based line and column, using
// token.FileSet's standard Go position conventions.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span is an inclusive-start, exclusive-end source range.
type Span struct {
	Filename string
	Start    Position
	End      Position
}

// String formats a source span for diagnostics without requiring source text.
func (s Span) String() string {
	location := fmt.Sprintf("%d:%d", s.Start.Line, s.Start.Column)
	if s.Filename != "" {
		location = s.Filename + ":" + location
	}
	if s.Start.Offset == s.End.Offset {
		return location
	}
	return fmt.Sprintf("%s-%d:%d", location, s.End.Line, s.End.Column)
}

// Fact is a deterministic semantic relation produced from one source reference.
type Fact struct {
	Subject  Identity
	Relation Relation
	Object   Identity
	Span     Span
	Origin   ObservationOrigin
}
