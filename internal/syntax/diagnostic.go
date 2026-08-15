package syntax

import (
	"sort"
	"strings"
)

// Severity classifies a diagnostic. The syntax layer currently emits errors;
// warnings remain part of the API so future syntax extensions do not need an
// incompatible result type.
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

// DiagnosticCode is a stable machine-readable diagnostic identifier.
type DiagnosticCode string

const (
	DiagUnexpectedCharacter   DiagnosticCode = "lex.unexpected-character"
	DiagUnterminatedComment   DiagnosticCode = "lex.unterminated-comment"
	DiagUnterminatedString    DiagnosticCode = "lex.unterminated-string"
	DiagInvalidEscape         DiagnosticCode = "lex.invalid-escape"
	DiagInvalidUTF8           DiagnosticCode = "lex.invalid-utf8"
	DiagExpectedPackage       DiagnosticCode = "parse.expected-package"
	DiagExpectedNamespace     DiagnosticCode = "parse.expected-namespace"
	DiagExpectedIdentifier    DiagnosticCode = "parse.expected-identifier"
	DiagExpectedID            DiagnosticCode = "parse.expected-id"
	DiagExpectedString        DiagnosticCode = "parse.expected-string"
	DiagExpectedLeftParen     DiagnosticCode = "parse.expected-left-paren"
	DiagExpectedRightParen    DiagnosticCode = "parse.expected-right-paren"
	DiagExpectedComma         DiagnosticCode = "parse.expected-comma"
	DiagExpectedArrow         DiagnosticCode = "parse.expected-arrow"
	DiagExpectedResult        DiagnosticCode = "parse.expected-result"
	DiagUnexpectedDeclaration DiagnosticCode = "parse.unexpected-token"
)

const (
	DiagEntityFieldsDeferred      DiagnosticCode = "parse.entity-fields-deferred"
	DiagEntityFieldsConfiguration DiagnosticCode = "parse.entity-fields-configuration"
)

// Diagnostic describes a recoverable lexical or syntactic problem.
type Diagnostic struct {
	Severity Severity
	Code     DiagnosticCode
	Message  string
	Span     Span
}

func (d Diagnostic) Error() string {
	return d.String()
}

// String formats a diagnostic deterministically.
func (d Diagnostic) String() string {
	return d.Span.String() + ": " + d.Severity.String() + " " + string(d.Code) + ": " + d.Message
}

// Diagnostics is an ordered list of diagnostics. Lexer diagnostics precede
// parser diagnostics, and diagnostics produced within each phase follow
// source order.
type Diagnostics []Diagnostic

// HasErrors reports whether at least one error was emitted.
func (d Diagnostics) HasErrors() bool {
	for _, diagnostic := range d {
		if diagnostic.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Errors returns a copy containing only error diagnostics.
func (d Diagnostics) Errors() Diagnostics {
	errors := make(Diagnostics, 0, len(d))
	for _, diagnostic := range d {
		if diagnostic.Severity == SeverityError {
			errors = append(errors, diagnostic)
		}
	}
	return errors
}

// SortBySpan returns a copy sorted by the canonical diagnostic key: filename,
// start offset, end offset, severity, code, then message. Parse results
// preserve phase order; callers that need one globally ordered view can use
// this method.
func (d Diagnostics) SortBySpan() Diagnostics {
	result := append(Diagnostics(nil), d...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Span.Filename != result[j].Span.Filename {
			return result[i].Span.Filename < result[j].Span.Filename
		}
		return diagnosticLess(result[i], result[j])
	})
	return result
}

func diagnosticLess(left, right Diagnostic) bool {
	if left.Span.Start.Offset != right.Span.Start.Offset {
		return left.Span.Start.Offset < right.Span.Start.Offset
	}
	if left.Span.End.Offset != right.Span.End.Offset {
		return left.Span.End.Offset < right.Span.End.Offset
	}
	if left.Severity != right.Severity {
		return left.Severity < right.Severity
	}
	if left.Code != right.Code {
		return left.Code < right.Code
	}
	return left.Message < right.Message
}

// Error returns all diagnostics as one deterministic error value. It returns
// nil when there are no errors, even if warnings are present.
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
