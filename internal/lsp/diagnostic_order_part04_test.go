package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func mapDiagnosticsForURI(uri, source string, diagnostics syntax.Diagnostics) ([]Diagnostic, error) {
	result := make([]Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		mapped, err := syntaxDiagnostic(source, diagnostic)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return canonicalDiagnosticOrder(uri, source, result), nil
}
func expectedAdapterReplay(t *testing.T, source string, raw syntax.Diagnostics) []Diagnostic {
	t.Helper()
	result, err := adaptSyntaxResult("cross-view.gooo", source, &syntax.File{}, raw)
	if err != nil {
		t.Fatal(err)
	}
	return result.Diagnostics
}

type diagnosticPair struct {
	Code    string
	Message string
}

func assertDiagnosticPairs(t *testing.T, diagnostics []Diagnostic, want []diagnosticPair) {
	t.Helper()
	if len(diagnostics) != len(want) {
		t.Fatalf("diagnostics = %#v, want %v", diagnostics, want)
	}
	for index, pair := range want {
		got := diagnosticPair{Code: diagnostics[index].Code, Message: diagnostics[index].Message}
		if got != pair {
			t.Fatalf("diagnostic[%d] = %#v, want %#v", index, got, pair)
		}
	}
}
func syntaxDiagnosticAt(filename string, start, end int, severity syntax.Severity, code, message string) syntax.Diagnostic {
	return syntax.Diagnostic{
		Severity: severity,
		Code:     syntax.DiagnosticCode(code),
		Message:  message,
		Span: syntax.Span{
			Filename: filename,
			Start:    syntax.Position{Offset: start},
			End:      syntax.Position{Offset: end},
		},
	}
}
func mapDiagnosticOrder(t *testing.T, source string, raw syntax.Diagnostics) []Diagnostic {
	t.Helper()
	mapped := make([]Diagnostic, 0, len(raw))
	for _, diagnostic := range raw {
		value, err := syntaxDiagnostic(source, diagnostic)
		if err != nil {
			t.Fatal(err)
		}
		mapped = append(mapped, value)
	}
	return canonicalDiagnosticOrder("replay.gooo", source, mapped)
}
