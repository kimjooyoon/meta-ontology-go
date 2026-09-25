package syntax

import (
	"testing"
)

func TestParserRootAndEOFDiagnosticsUseLexerBoundary(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantLine   int
		wantColumn int
		wantDiags  int
	}{
		{name: "empty", source: "", wantLine: 1, wantColumn: 1, wantDiags: 2},
		{name: "CRLF", source: "package p\r\n", wantLine: 2, wantColumn: 1, wantDiags: 1},
		{name: "Unicode", source: "package 도메인\r\n", wantLine: 2, wantColumn: 1, wantDiags: 1},
		{name: "trailing bytes", source: "package p\nnamespace n\t", wantLine: 2, wantColumn: 13, wantDiags: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := test.source
			file, diagnostics := ParseFile("boundary.gooo", test.source)
			tokens, lexerDiagnostics := LexFile("boundary.gooo", test.source)
			eof := tokens[len(tokens)-1].Span
			if file.Span.End != eof.End {
				t.Fatalf("root end = %#v, lexer EOF = %#v", file.Span.End, eof.End)
			}
			if file.Span.End.Offset != len(test.source) || file.Span.End.Line != test.wantLine || file.Span.End.Column != test.wantColumn {
				t.Fatalf("root end = %#v, want EOF %d:%d at byte %d", file.Span.End, test.wantLine, test.wantColumn, len(test.source))
			}
			if len(diagnostics) != test.wantDiags || len(lexerDiagnostics) != 0 {
				t.Fatalf("diagnostics = %v, lexer diagnostics = %v", diagnostics, lexerDiagnostics)
			}
			for _, diagnostic := range diagnostics {
				if diagnostic.Code == DiagExpectedPackage || diagnostic.Code == DiagExpectedNamespace {
					if diagnostic.Span.Start != eof.Start || diagnostic.Span.End != eof.End {
						t.Fatalf("EOF diagnostic span = %#v, want %#v", diagnostic.Span, eof)
					}
				}
			}
			if test.source != original {
				t.Fatal("parser modified source input")
			}
		})
	}
}
