// Package formatter contains the parser-neutral formatting boundary for .gooo.
//
// The package intentionally does not import internal/syntax. A syntax package
// can adapt its AST through ASTAdapter without coupling the formatter to one
// provisional tree shape.
package formatter

import (
	"fmt"
	"strings"
)

// Severity classifies a formatter diagnostic.
type Severity uint8

const (
	SeverityError Severity = iota
	SeverityWarning
)

func (s Severity) String() string {
	if s == SeverityWarning {
		return "warning"
	}
	return "error"
}

// DiagnosticCode is a stable machine-readable formatter diagnostic code.
type DiagnosticCode string

const (
	CodeMissingAST          DiagnosticCode = "formatter.missing-ast"
	CodeMissingAdapter      DiagnosticCode = "formatter.missing-adapter"
	CodeAdapterReturnedNil  DiagnosticCode = "formatter.adapter-returned-nil"
	CodeInvalidDocument     DiagnosticCode = "formatter.invalid-document"
	CodeUnsupportedIdentity DiagnosticCode = "formatter.unsupported-identity"
	CodeUnsupportedSyntax   DiagnosticCode = "formatter.unsupported-syntax"
)

// Position is a one-based source position. Formatter diagnostics do not need
// to know the concrete syntax package's span type.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span is a half-open source range. A zero Span is valid for document-level
// diagnostics such as a missing AST.
type Span struct {
	Filename string
	Start    Position
	End      Position
}

// Diagnostic describes a recoverable formatting problem.
type Diagnostic struct {
	Severity Severity
	Code     DiagnosticCode
	Message  string
	Span     Span
}

func (d Diagnostic) Error() string { return d.String() }

// String returns a deterministic diagnostic representation.
func (d Diagnostic) String() string {
	location := fmt.Sprintf("%d:%d", d.Span.Start.Line, d.Span.Start.Column)
	if d.Span.Filename != "" {
		location = d.Span.Filename + ":" + location
	}
	return fmt.Sprintf("%s: %s %s: %s", location, d.Severity, d.Code, d.Message)
}

// Diagnostics is an ordered list of formatter diagnostics.
type Diagnostics []Diagnostic

// HasErrors reports whether formatting should be considered unsuccessful.
func (d Diagnostics) HasErrors() bool {
	for _, diagnostic := range d {
		if diagnostic.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Error joins error diagnostics while ignoring warnings.
func (d Diagnostics) Error() error {
	if !d.HasErrors() {
		return nil
	}
	lines := make([]string, 0, len(d))
	for _, diagnostic := range d {
		if diagnostic.Severity == SeverityError {
			lines = append(lines, diagnostic.String())
		}
	}
	return diagnosticError(strings.Join(lines, "\n"))
}

type diagnosticError string

func (e diagnosticError) Error() string { return string(e) }
