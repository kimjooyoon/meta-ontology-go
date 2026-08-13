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

func TestAdapterReplayUsesAllEqualStartTieBreaks(t *testing.T) {
	source := "0123456789012345678901234567890123456789"
	raw := syntax.Diagnostics{
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityWarning, "a.code", "warning"),
		syntaxDiagnosticAt("cross-view.gooo", 10, 12, syntax.SeverityError, "early", "early"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 22, syntax.SeverityError, "z.code", "short-end"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityError, "z.code", "z-message"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityError, "a.code", "z-message"),
		syntaxDiagnosticAt("cross-view.gooo", 20, 23, syntax.SeverityError, "a.code", "a-message"),
	}
	want := expectedAdapterReplay(t, source, raw)
	for replay := 0; replay < 64; replay++ {
		permuted := rotateDiagnostics(raw, replay)
		result, err := adaptSyntaxResult("cross-view.gooo", source, &syntax.File{}, permuted)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(want, result.Diagnostics) {
			t.Fatalf("replay %d changed adapter order: want=%#v got=%#v", replay, want, result.Diagnostics)
		}
	}
	assertDiagnosticPairs(t, want, []diagnosticPair{
		{Code: "early", Message: "early"},
		{Code: "z.code", Message: "short-end"},
		{Code: "a.code", Message: "a-message"},
		{Code: "a.code", Message: "z-message"},
		{Code: "z.code", Message: "z-message"},
		{Code: "a.code", Message: "warning"},
	})
}

func TestCanonicalDiagnosticOrderSpecialSourceReplay(t *testing.T) {
	invalidUTF8 := string(append([]byte("α😀\r\nx "), 0xff, 'z'))
	cases := []struct {
		name        string
		source      string
		diagnostics syntax.Diagnostics
	}{
		{
			name:   "equal-span",
			source: "0123456789",
			diagnostics: syntax.Diagnostics{
				syntaxDiagnosticAt("equal.gooo", 4, 6, syntax.SeverityError, "a", "alpha"),
				syntaxDiagnosticAt("equal.gooo", 4, 6, syntax.SeverityError, "b", "beta"),
				syntaxDiagnosticAt("equal.gooo", 4, 6, syntax.SeverityWarning, "z", "warning"),
			},
		},
		{
			name:   "crlf",
			source: "package p\r\nnamespace n\r\n@",
			diagnostics: syntax.Diagnostics{
				syntaxDiagnosticAt("crlf.gooo", 11, 12, syntax.SeverityError, "line", "line"),
				syntaxDiagnosticAt("crlf.gooo", 24, 25, syntax.SeverityError, "late", "late"),
			},
		},
		{
			name:   "unicode",
			source: "α😀\r\n終わり",
			diagnostics: syntax.Diagnostics{
				syntaxDiagnosticAt("unicode.gooo", 0, 6, syntax.SeverityError, "a", "unicode"),
				syntaxDiagnosticAt("unicode.gooo", 6, 8, syntax.SeverityWarning, "z", "終わり"),
			},
		},
		{
			name:   "invalid-utf8",
			source: invalidUTF8,
			diagnostics: syntax.Diagnostics{
				syntaxDiagnosticAt("invalid.gooo", 6, 8, syntax.SeverityError, "crlf", "line"),
				syntaxDiagnosticAt("invalid.gooo", 10, 11, syntax.SeverityError, "invalid", "byte"),
			},
		},
	}
	for _, fixture := range cases {
		t.Run(fixture.name, func(t *testing.T) {
			wantRaw := append(syntax.Diagnostics(nil), fixture.diagnostics...)
			syntaxView := fixture.diagnostics.SortBySpan()
			if !reflect.DeepEqual(wantRaw, syntaxView) {
				t.Fatalf("syntax SortBySpan changed canonical fixture: want=%#v got=%#v", wantRaw, syntaxView)
			}
			want, err := mapDiagnosticsForURI(fixture.name, fixture.source, syntaxView)
			if err != nil {
				t.Fatal(err)
			}
			originalSource := []byte(fixture.source)
			originalDiagnostics := append(syntax.Diagnostics(nil), fixture.diagnostics...)
			for replay := 0; replay < 64; replay++ {
				permuted := rotateDiagnostics(fixture.diagnostics, replay)
				result, err := adaptSyntaxResult(fixture.name, fixture.source, &syntax.File{}, permuted)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(want, result.Diagnostics) {
					t.Fatalf("replay %d changed cross-view order: want=%#v got=%#v", replay, want, result.Diagnostics)
				}
				if !reflect.DeepEqual(originalDiagnostics, fixture.diagnostics) || string(originalSource) != fixture.source {
					t.Fatalf("replay %d mutated diagnostic/source input", replay)
				}
			}
		})
	}
}

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
