package syntax

import (
	"sort"
	"strings"
)

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
