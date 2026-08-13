package bidir

import (
	"fmt"
	"strings"
)

// SourceSpan is the dependency-free provenance boundary used by adapters.
type SourceSpan struct {
	File        string
	Start       int
	End         int
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

// Validate checks the structural invariants of a source span. The zero value
// is valid and means that no source evidence was supplied.
func (s SourceSpan) Validate() error {
	if s.Start < 0 || s.End < 0 {
		return fmt.Errorf("source span offsets must be non-negative")
	}
	if s.End < s.Start {
		return fmt.Errorf("source span end precedes start")
	}
	if s.StartLine < 0 || s.StartColumn < 0 || s.EndLine < 0 || s.EndColumn < 0 {
		return fmt.Errorf("source span positions must be non-negative")
	}
	if s.StartLine > 0 && s.EndLine > 0 {
		if s.EndLine < s.StartLine {
			return fmt.Errorf("source span end line precedes start line")
		}
		if s.EndLine == s.StartLine && s.StartColumn > 0 && s.EndColumn > 0 && s.EndColumn < s.StartColumn {
			return fmt.Errorf("source span end column precedes start column")
		}
	}
	return nil
}

// Valid reports whether the span carries structurally valid source evidence.
func (s SourceSpan) Valid() bool {
	return s.Validate() == nil && (strings.TrimSpace(s.File) != "" || s.Start != 0 || s.End != 0 || s.StartLine != 0 || s.EndLine != 0)
}
