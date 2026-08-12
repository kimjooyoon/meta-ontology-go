package lsp

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestCanonicalDiagnosticOrderReplaysEqualStarts(t *testing.T) {
	source := "0123456789012345678901234567890123456789"
	raw := syntax.Diagnostics{
		syntaxDiagnosticAt("replay.gooo", 20, 24, syntax.SeverityWarning, "z.code", "warning"),
		syntaxDiagnosticAt("replay.gooo", 10, 12, syntax.SeverityError, "early", "early"),
		syntaxDiagnosticAt("replay.gooo", 20, 23, syntax.SeverityError, "b.code", "short"),
		syntaxDiagnosticAt("replay.gooo", 20, 23, syntax.SeverityError, "a.code", "alpha"),
		syntaxDiagnosticAt("replay.gooo", 20, 23, syntax.SeverityError, "a.code", "beta"),
		syntaxDiagnosticAt("replay.gooo", 30, 31, syntax.SeverityError, "late", "late"),
	}
	want := mapDiagnosticOrder(t, source, raw)
	for replay := 0; replay < 32; replay++ {
		permuted := rotateDiagnostics(raw, replay)
		got := mapDiagnosticOrder(t, source, permuted)
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("replay %d changed order: want=%#v got=%#v", replay, want, got)
		}
	}
	assertDiagnosticCodes(t, want, []string{"early", "a.code", "a.code", "b.code", "z.code", "late"})
	if want[1].Message != "alpha" || want[2].Message != "beta" {
		t.Fatalf("equal-span message order = %#v", want[1:3])
	}
}

func TestCanonicalDiagnosticOrderMatchesSyntaxSourceReplay(t *testing.T) {
	source := "0123456789012345678901234567890123456789"
	raw := syntax.Diagnostics{
		syntaxDiagnosticAt("replay.gooo", 20, 23, syntax.SeverityError, "z.code", "z"),
		syntaxDiagnosticAt("replay.gooo", 10, 11, syntax.SeverityError, "early", "early"),
		syntaxDiagnosticAt("replay.gooo", 20, 22, syntax.SeverityError, "a.code", "a"),
	}
	syntaxView := raw.SortBySpan()
	want := mapDiagnosticOrder(t, source, syntaxView)
	for replay := 0; replay < 16; replay++ {
		mapped := make([]Diagnostic, 0, len(raw))
		for _, diagnostic := range rotateDiagnostics(raw, replay) {
			value, err := syntaxDiagnostic(source, diagnostic)
			if err != nil {
				t.Fatal(err)
			}
			mapped = append(mapped, value)
		}
		got := canonicalDiagnosticOrder("replay.gooo", source, mapped)
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("cross-view replay %d changed order: want=%#v got=%#v", replay, want, got)
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

func rotateDiagnostics(values syntax.Diagnostics, shift int) syntax.Diagnostics {
	result := append(syntax.Diagnostics(nil), values...)
	if len(result) == 0 {
		return result
	}
	shift %= len(result)
	return append(result[shift:], result[:shift]...)
}

func assertDiagnosticCodes(t *testing.T, diagnostics []Diagnostic, want []string) {
	t.Helper()
	if len(diagnostics) != len(want) {
		t.Fatalf("diagnostics = %#v, want %v", diagnostics, want)
	}
	for index, code := range want {
		if diagnostics[index].Code != code {
			t.Fatalf("diagnostics = %#v, want %v", diagnostics, want)
		}
	}
}
