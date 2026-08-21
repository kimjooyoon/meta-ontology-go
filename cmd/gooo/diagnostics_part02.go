package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"sort"
)

// canonicalDiagnostics provides the CLI's cross-view ordering contract. The
// syntax package preserves parser order for equal start positions, which is
// useful to syntax consumers but is not enough for a serialized CLI view:
// LSP and future diagnostics producers can report different phases at the
// same source location. Sort every observable field after the source span so
// replaying the same set in another insertion order produces identical bytes.
func canonicalDiagnostics(diagnostics syntax.Diagnostics) syntax.Diagnostics {
	result := append(syntax.Diagnostics(nil), diagnostics...)
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Span.Filename != right.Span.Filename {
			return left.Span.Filename < right.Span.Filename
		}
		if left.Span.Start.Offset != right.Span.Start.Offset {
			return left.Span.Start.Offset < right.Span.Start.Offset
		}
		if left.Span.Start.Line != right.Span.Start.Line {
			return left.Span.Start.Line < right.Span.Start.Line
		}
		if left.Span.Start.Column != right.Span.Start.Column {
			return left.Span.Start.Column < right.Span.Start.Column
		}
		if left.Span.End.Offset != right.Span.End.Offset {
			return left.Span.End.Offset < right.Span.End.Offset
		}
		if left.Span.End.Line != right.Span.End.Line {
			return left.Span.End.Line < right.Span.End.Line
		}
		if left.Span.End.Column != right.Span.End.Column {
			return left.Span.End.Column < right.Span.End.Column
		}
		if left.Severity != right.Severity {
			return left.Severity < right.Severity
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		return left.Message < right.Message
	})
	return result
}
