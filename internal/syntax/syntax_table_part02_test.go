package syntax

import (
	"reflect"
	"testing"
)

func TestLexSpansAndDeterminism(t *testing.T) {
	source := "package p\nnamespace n\n"
	firstTokens, firstDiagnostics := LexFile("billing.gooo", source)
	secondTokens, secondDiagnostics := LexFile("billing.gooo", source)
	if !reflect.DeepEqual(firstTokens, secondTokens) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("lexing the same source was not deterministic")
	}

	if got, want := firstTokens[0].Span, (Span{Filename: "billing.gooo", Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 7, Line: 1, Column: 8}}); got != want {
		t.Fatalf("package span = %#v, want %#v", got, want)
	}
	if got, want := firstTokens[2].Span.Start, (Position{Offset: 10, Line: 2, Column: 1}); got != want {
		t.Fatalf("namespace start = %#v, want %#v", got, want)
	}
	if got, want := firstTokens[len(firstTokens)-1].Span.Start, (Position{Offset: len(source), Line: 3, Column: 1}); got != want {
		t.Fatalf("EOF position = %#v, want %#v", got, want)
	}
}
func TestLexDiagnosticsTable(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   DiagnosticCode
	}{
		{name: "bad character", source: "entity A id \"x\" @", code: DiagUnexpectedCharacter},
		{name: "unterminated block comment", source: "/* comment", code: DiagUnterminatedComment},
		{name: "unterminated string", source: "entity A id \"x", code: DiagUnterminatedString},
		{name: "invalid escape", source: "entity A id \"\\q\"", code: DiagInvalidEscape},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Lex(test.source)
			if len(diagnostics) != 1 || diagnostics[0].Code != test.code {
				t.Fatalf("diagnostics = %#v, want one %q", diagnostics, test.code)
			}
		})
	}
}
