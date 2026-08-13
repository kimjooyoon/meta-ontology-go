package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func TestMalformedEscapeRecoveryPreservesCRLFSpans(t *testing.T) {
	source := "package p\r\nnamespace n\r\nentity A id \"prefix\\\r\n"
	original := source
	file, diagnostics := ParseFile("malformed.gooo", source)
	tokens, lexerDiagnostics := LexFile("malformed.gooo", source)
	if len(lexerDiagnostics) != len(diagnostics) || len(diagnostics) != 1 {
		t.Fatalf("parse diagnostics = %#v, lexer diagnostics = %#v", diagnostics, lexerDiagnostics)
	}
	if diagnostics[0].Code != DiagUnterminatedString {
		t.Fatalf("diagnostic code = %q, want %q", diagnostics[0].Code, DiagUnterminatedString)
	}
	if file.Span.End != tokens[len(tokens)-1].Span.End {
		t.Fatalf("root end = %#v, lexer EOF = %#v", file.Span.End, tokens[len(tokens)-1].Span.End)
	}
	if file.Span.End.Offset != len(source) || file.Span.End.Line != 4 || file.Span.End.Column != 1 {
		t.Fatalf("root end = %#v, want EOF at byte %d, line 4, column 1", file.Span.End, len(source))
	}
	backslash := strings.IndexByte(source, '\\')
	wantSpan := Span{Filename: "malformed.gooo", Start: Position{Offset: backslash, Line: 3, Column: 20}, End: Position{Offset: backslash + 1, Line: 3, Column: 21}}
	if diagnostics[0].Span != wantSpan {
		t.Fatalf("escape diagnostic span = %#v, want %#v", diagnostics[0].Span, wantSpan)
	}
	if source != original {
		t.Fatal("parser modified source input")
	}
}

func TestMalformedSourcesRejectFormattingWithoutMutation(t *testing.T) {
	invalidUTF8 := string([]byte{'x', 0xff, 'y'})
	sources := []string{
		"package p\r\nnamespace n\r\nentity A id \"prefix\\\r\n",
		quotedIDSource("A", invalidUTF8),
		quotedIDSource("A", `\u12xz`),
		"package p\nnamespace n\nentity A id \"unterminated",
	}
	for _, source := range sources {
		original := source
		file, firstDiagnostics := ParseFile("malformed.gooo", source)
		formatted, formatDiagnostics, err := FormatSource("malformed.gooo", source)
		secondFile, secondDiagnostics := ParseFile("malformed.gooo", source)
		if err == nil || formatted != "" || len(formatDiagnostics) == 0 {
			t.Fatalf("malformed format result = %q, %#v, %v", formatted, formatDiagnostics, err)
		}
		if !reflect.DeepEqual(file, secondFile) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
			t.Fatalf("repeated malformed parse differed for %q", source)
		}
		if source != original {
			t.Fatal("formatting modified source input")
		}
	}
}

func TestMalformedUnicodeEscapeStopsBeforeNewline(t *testing.T) {
	tests := []struct {
		name    string
		newline string
	}{
		{name: "LF", newline: "\n"},
		{name: "CRLF", newline: "\r\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := "package p\r\nnamespace n\r\nentity A id \"prefix\\u12" + test.newline
			file, diagnostics := ParseFile("unicode-escape.gooo", source)
			newlineOffset := strings.LastIndex(source, test.newline)
			if file.Span.End.Offset != len(source) || len(diagnostics) != 2 {
				t.Fatalf("file = %#v, diagnostics = %#v", file.Span, diagnostics)
			}
			if diagnostics[0].Code != DiagInvalidEscape || diagnostics[1].Code != DiagUnterminatedString {
				t.Fatalf("diagnostic codes = %#v", diagnostics)
			}
			if diagnostics[0].Span.End.Offset != newlineOffset || diagnostics[1].Span.End.Offset != newlineOffset {
				t.Fatalf("diagnostic spans crossed newline: %#v", diagnostics)
			}
			formatted, formatDiagnostics, err := FormatSource("unicode-escape.gooo", source)
			if err == nil || formatted != "" || !reflect.DeepEqual(diagnostics, formatDiagnostics) {
				t.Fatalf("format result = %q, %#v, %v", formatted, formatDiagnostics, err)
			}
			secondFile, secondDiagnostics := ParseFile("unicode-escape.gooo", source)
			if !reflect.DeepEqual(file, secondFile) || !reflect.DeepEqual(diagnostics, secondDiagnostics) {
				t.Fatal("malformed Unicode escape recovery was not deterministic")
			}
		})
	}
}

func TestMalformedUnicodeEscapePreservesClosingQuote(t *testing.T) {
	source := quotedIDSource("A", `\u12`)
	original := source
	file, diagnostics := ParseFile("unicode-quote.gooo", source)
	tokens, lexerDiagnostics := LexFile("unicode-quote.gooo", source)
	if len(diagnostics) != 1 || len(lexerDiagnostics) != 1 || diagnostics[0].Code != DiagInvalidEscape {
		t.Fatalf("diagnostics = %#v, lexer diagnostics = %#v", diagnostics, lexerDiagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	quoteOffset := strings.LastIndexByte(source, '"')
	if entity.IDSpan.End.Offset != quoteOffset+1 || file.Span.End != tokens[len(tokens)-1].Span.End {
		t.Fatalf("string/root spans = %#v, %#v", entity.IDSpan, file.Span)
	}
	formatted, formatDiagnostics, err := FormatSource("unicode-quote.gooo", source)
	if err == nil || formatted != "" || !reflect.DeepEqual(diagnostics, formatDiagnostics) {
		t.Fatalf("format result = %q, %#v, %v", formatted, formatDiagnostics, err)
	}
	secondFile, secondDiagnostics := ParseFile("unicode-quote.gooo", source)
	if !reflect.DeepEqual(file, secondFile) || !reflect.DeepEqual(diagnostics, secondDiagnostics) || source != original {
		t.Fatal("malformed Unicode quote recovery was not deterministic or mutated input")
	}
}
