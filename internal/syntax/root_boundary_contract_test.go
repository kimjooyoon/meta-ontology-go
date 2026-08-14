package syntax

import (
	"reflect"
	"strings"
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

func TestParserRecoveryAndDiagnosticOrderingDoNotMutateInputs(t *testing.T) {
	invalid := string([]byte{0xff})
	source := "package 도메인\r\nnamespace 도메인\r\nentity 注文 id \"앞" + invalid + "뒤\"\r\n"
	originalSource := source
	parser := NewParserFile("immutable.gooo", source)
	file, diagnostics := parser.Parse()
	repeatedFile, repeatedDiagnostics := parser.Parse()
	originalDiagnostics := append(Diagnostics(nil), diagnostics...)
	firstSorted := diagnostics.SortBySpan()
	secondSorted := diagnostics.SortBySpan()
	invalidOffset := strings.IndexByte(source, 0xff)

	if source != originalSource {
		t.Fatal("parser modified source input")
	}
	if !reflect.DeepEqual(diagnostics, originalDiagnostics) {
		t.Fatal("sorting modified diagnostic input")
	}
	if !reflect.DeepEqual(firstSorted, secondSorted) {
		t.Fatalf("diagnostic ordering was not deterministic: %#v vs %#v", firstSorted, secondSorted)
	}
	if !reflect.DeepEqual(file, repeatedFile) || !reflect.DeepEqual(diagnostics, repeatedDiagnostics) {
		t.Fatal("repeated recovery changed the parsed result")
	}
	if file.Span.End.Offset != len(source) || file.Span.End.Line != 4 || file.Span.End.Column != 1 {
		t.Fatalf("recovered root span = %#v", file.Span)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagInvalidUTF8 {
		t.Fatalf("malformed UTF-8 diagnostics = %#v", diagnostics)
	}
	invalidSpan := diagnostics[0].Span
	if invalidSpan.Start.Offset != invalidOffset || invalidSpan.End.Offset != invalidOffset+1 || invalidSpan.Start.Line != 3 || invalidSpan.Start.Column != 16 {
		t.Fatalf("invalid UTF-8 span = %#v, want byte %d at 3:16", invalidSpan, invalidOffset)
	}
	for index, diagnostic := range firstSorted {
		if index == 0 || diagnostic.Span.Start.Offset >= firstSorted[index-1].Span.Start.Offset {
			continue
		}
		t.Fatalf("diagnostics were not ordered by byte offset: %#v", firstSorted)
	}
}
