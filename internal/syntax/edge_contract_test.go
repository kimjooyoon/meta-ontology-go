package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func TestEmptySourceHasStableEOFAndRootSpan(t *testing.T) {
	tokens, lexDiagnostics := LexFile("empty.gooo", "")
	if len(lexDiagnostics) != 0 || len(tokens) != 1 || tokens[0].Kind != TokenEOF {
		t.Fatalf("empty lex result = %#v, %v", tokens, lexDiagnostics)
	}
	eof := tokens[0].Span
	if eof.Start != (Position{Offset: 0, Line: 1, Column: 1}) || eof.End != eof.Start {
		t.Fatalf("empty EOF span = %#v", eof)
	}
	file, parseDiagnostics := ParseFile("empty.gooo", "")
	if len(parseDiagnostics) != 2 || file.Span.Start != eof.Start || file.Span.End != eof.End {
		t.Fatalf("empty parse result = %#v, %v", file, parseDiagnostics)
	}
}

func TestRootSpanEndsAtEOFAfterFullParse(t *testing.T) {
	source := "package p\nnamespace n"
	file, diagnostics := ParseFile("eof.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	tokens, _ := LexFile("eof.gooo", source)
	eof := tokens[len(tokens)-1].Span
	if file.Span.End != eof.End || file.Span.End.Offset != len(source) || file.Span.End.Line != 2 || file.Span.End.Column != 12 {
		t.Fatalf("root span end = %#v, EOF = %#v", file.Span.End, eof.End)
	}
}

func TestCRLFAndUnicodePositionsUseBytesAndRunes(t *testing.T) {
	source := "package 도메인\r\nnamespace 도메인\r\nentity 注文 id \"urn:注文\"\r\n"
	file, diagnostics := ParseFile("unicode.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	if entity.NameSpan.Start.Line != 3 || entity.NameSpan.Start.Column != 8 || entity.NameSpan.End.Column != 10 {
		t.Fatalf("unicode entity span = %#v", entity.NameSpan)
	}
	if entity.NameSpan.Len() != len("注文") || file.Span.End != (Position{Offset: len(source), Line: 4, Column: 1}) {
		t.Fatalf("unicode byte/root spans = %#v, %#v", entity.NameSpan, file.Span)
	}
}

func TestQuotedStringRejectsMalformedUTF8WithByteSpans(t *testing.T) {
	invalid := string([]byte{0xff, 0xfe})
	source := "package p\nnamespace n\nentity A id \"" + invalid + "\""
	_, diagnostics := ParseFile("invalid-utf8.gooo", source)
	if len(diagnostics) != 2 || diagnostics[0].Code != DiagInvalidUTF8 || diagnostics[1].Code != DiagInvalidUTF8 {
		t.Fatalf("malformed UTF-8 diagnostics = %#v", diagnostics)
	}
	firstOffset := strings.IndexByte(source, 0xff)
	for index, diagnostic := range diagnostics {
		wantOffset := firstOffset + index
		if diagnostic.Span.Start.Offset != wantOffset || diagnostic.Span.End.Offset != wantOffset+1 || diagnostic.Span.Len() != 1 {
			t.Fatalf("diagnostic %d span = %#v, want byte %d", index, diagnostic.Span, wantOffset)
		}
	}
}

func TestDiagnosticsSortBySpanHasDeterministicTieBreaks(t *testing.T) {
	span := Span{Filename: "same.gooo", Start: Position{Offset: 4, Line: 1, Column: 5}, End: Position{Offset: 5, Line: 1, Column: 6}}
	shorter := span
	shorter.End.Offset = 4
	input := Diagnostics{
		{Code: "z-code", Message: "z", Span: span},
		{Code: "a-code", Message: "a", Span: span, Severity: SeverityWarning},
		{Code: "middle-code", Message: "middle", Span: shorter},
	}
	sorted := input.SortBySpan()
	wantCodes := []DiagnosticCode{"middle-code", "z-code", "a-code"}
	gotCodes := make([]DiagnosticCode, len(sorted))
	for index, diagnostic := range sorted {
		gotCodes[index] = diagnostic.Code
	}
	if !reflect.DeepEqual(gotCodes, wantCodes) || input[0].Code != "z-code" {
		t.Fatalf("sorted codes = %v, input = %v", gotCodes, input)
	}
}
