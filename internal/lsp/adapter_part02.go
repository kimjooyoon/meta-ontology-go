package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"sort"
)

// canonicalDiagnosticOrder preserves syntax source order and adds only the
// LSP view's deterministic tie-breaks for diagnostics sharing a start.
func canonicalDiagnosticOrder(uri, source string, diagnostics []Diagnostic) []Diagnostic {
	result := append([]Diagnostic(nil), diagnostics...)
	for index := range result {
		if result[index].filename == "" {
			result[index].filename = uri
		}
		if !result[index].spanned {
			result[index].start, _ = PositionToOffset(source, result[index].Range.Start)
			result[index].end, _ = PositionToOffset(source, result[index].Range.End)
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.filename != second.filename {
			return first.filename < second.filename
		}
		if first.start != second.start {
			return first.start < second.start
		}
		if first.end != second.end {
			return first.end < second.end
		}
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		if first.Range.End != second.Range.End {
			return positionLess(first.Range.End, second.Range.End)
		}
		if first.Severity != second.Severity {
			return first.Severity < second.Severity
		}
		if first.Code != second.Code {
			return first.Code < second.Code
		}
		return first.Message < second.Message
	})
	return result
}
func positionLess(first, second Position) bool {
	if first.Line != second.Line {
		return first.Line < second.Line
	}
	return first.Character < second.Character
}
func syntaxDiagnostic(source string, diagnostic syntax.Diagnostic) (Diagnostic, error) {
	rangeValue, err := syntaxRange(source, diagnostic.Span)
	if err != nil {
		return Diagnostic{}, err
	}
	severity := DiagnosticError
	if diagnostic.Severity == syntax.SeverityWarning {
		severity = DiagnosticWarning
	}
	return Diagnostic{
		Range: rangeValue, Severity: severity, Code: string(diagnostic.Code),
		Source: "gooo", Message: diagnostic.Message, filename: diagnostic.Span.Filename,
		start: diagnostic.Span.Start.Offset, end: diagnostic.Span.End.Offset, spanned: true,
	}, nil
}
