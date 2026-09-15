package main

import (
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
)

func writeJSONReport(writer io.Writer, report jsonReport) error {
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}
	if len(payload)+1 > maxDiagnosticBytes {
		return errDiagnosticLimit
	}
	_, err = fmt.Fprintf(writer, "%s\n", payload)
	return err
}
func sortedSyntaxDiagnostics(diagnostics syntax.Diagnostics) syntax.Diagnostics {
	return canonicalDiagnostics(diagnostics)
}
func syntaxCLIDiagnostics(diagnostics syntax.Diagnostics) []cliDiagnostic {
	sorted := sortedSyntaxDiagnostics(diagnostics)
	result := make([]cliDiagnostic, 0, len(sorted))
	for _, diagnostic := range sorted {
		result = append(result, cliDiagnostic{
			Severity: diagnostic.Severity.String(),
			Code:     string(diagnostic.Code),
			Message:  diagnostic.Message,
			Span:     cliSpanFromSyntax(diagnostic.Span),
		})
	}
	return result
}
func cliSpanFromSyntax(span syntax.Span) cliSpan {
	return cliSpan{
		File:  span.Filename,
		Start: cliPosition{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:   cliPosition{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}
func syntaxFileSpan(file *syntax.File) syntax.Span {
	if file == nil {
		return syntax.Span{}
	}
	return file.Span
}
func errorCLIDiagnostic(filename, code, message string, span syntax.Span) cliDiagnostic {
	if span.Filename == "" {
		span.Filename = filename
	}
	return cliDiagnostic{Severity: "error", Code: code, Message: message, Span: cliSpanFromSyntax(span)}
}
func printSyntaxDiagnostics(stderr io.Writer, diagnostics syntax.Diagnostics) error {
	for _, diagnostic := range sortedSyntaxDiagnostics(diagnostics) {
		if _, err := fmt.Fprintln(stderr, diagnostic.String()); err != nil {
			return err
		}
	}
	return nil
}
