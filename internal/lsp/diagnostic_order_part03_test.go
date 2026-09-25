package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

func TestCanonicalDiagnosticOrderSpecialSourceReplay(t *testing.T) {
	invalidUTF8 := string(append([]byte("α😀\r\nx "), 0xff, 'z'))
	cases := []struct {
		name        string
		source      string
		diagnostics syntax.Diagnostics
	}{{name: "equal-span", source: "0123456789", diagnostics: syntax.Diagnostics{syntaxDiagnosticAt("equal.gooo", 4, 6, syntax.SeverityError, "a", "alpha"), syntaxDiagnosticAt("equal.gooo", 4, 6, syntax.SeverityError, "b", "beta"), syntaxDiagnosticAt("equal.gooo", 4, 6, syntax.SeverityWarning, "z", "warning")}}, {name: "crlf", source: "package p\r\nnamespace n\r\n@", diagnostics: syntax.Diagnostics{syntaxDiagnosticAt("crlf.gooo", 11, 12, syntax.SeverityError, "line", "line"), syntaxDiagnosticAt("crlf.gooo", 24, 25, syntax.SeverityError, "late", "late")}}, {name: "unicode", source: "α😀\r\n終わり", diagnostics: syntax.Diagnostics{syntaxDiagnosticAt("unicode.gooo", 0, 6, syntax.SeverityError, "a", "unicode"), syntaxDiagnosticAt("unicode.gooo", 6, 8, syntax.SeverityWarning, "z", "終わり")}}, {name: "invalid-utf8", source: invalidUTF8, diagnostics: syntax.Diagnostics{syntaxDiagnosticAt("invalid.gooo", 6, 8, syntax.SeverityError, "crlf", "line"), syntaxDiagnosticAt("invalid.gooo", 10, 11, syntax.SeverityError, "invalid", "byte")}}}
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
			for replay := range 64 {
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
