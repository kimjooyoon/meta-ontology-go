package syntax

import "testing"

func TestFileSpanFinalizesAtLexerEOF(t *testing.T) {
	invalid := string([]byte{0xff, 0xfe})
	tests := []struct {
		name   string
		source string
		line   int
		column int
		diags  int
		lex    int
	}{
		{name: "empty", source: "", line: 1, column: 1, diags: 2},
		{name: "trailing whitespace", source: "package p\nnamespace n \t \n", line: 3, column: 1},
		{name: "trailing comments", source: "package p\nnamespace n\n// trailing\r\n", line: 4, column: 1},
		{name: "CRLF unicode", source: "package 도메인\r\nnamespace 도메인\r\n", line: 3, column: 1},
		{name: "malformed UTF-8 recovery", source: "package p\nnamespace n\nentity A id \"" + invalid + "\"\r\n", line: 4, column: 1, diags: 2, lex: 2},
		{name: "quoted diagnostic and comment", source: "package p\nnamespace n\nentity A id \"x\\q\" // trailing\n", line: 4, column: 1, diags: 1, lex: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, diagnostics := ParseFile("eof.gooo", test.source)
			tokens, lexDiagnostics := LexFile("eof.gooo", test.source)
			if len(diagnostics) != test.diags || len(lexDiagnostics) != test.lex {
				t.Fatalf("diagnostics = %v, lexer diagnostics = %v, want %d/%d", diagnostics, lexDiagnostics, test.diags, test.lex)
			}
			eof := tokens[len(tokens)-1]
			if eof.Kind != TokenEOF {
				t.Fatalf("last token = %#v, want EOF", eof)
			}
			if file.Span.End != eof.Span.End || file.Span.End.Offset != len(test.source) || file.Span.End.Line != test.line || file.Span.End.Column != test.column {
				t.Fatalf("file end = %#v, lexer EOF = %#v", file.Span.End, eof.Span.End)
			}
		})
	}
}

func TestFileSpanEOFIsDeterministicAcrossRepeatedRecovery(t *testing.T) {
	invalid := string([]byte{0xff})
	source := "package 도메인\r\nnamespace 도메인\r\nentity 注文 id \"앞" + invalid + "뒤\"\r\n// end\r\n"
	firstFile, firstDiagnostics := ParseFile("repeat.gooo", source)
	secondFile, secondDiagnostics := ParseFile("repeat.gooo", source)
	if firstFile.Span != secondFile.Span || firstDiagnostics.Error().Error() != secondDiagnostics.Error().Error() {
		t.Fatalf("repeated recovery differed: %#v/%v vs %#v/%v", firstFile.Span, firstDiagnostics, secondFile.Span, secondDiagnostics)
	}
	if firstFile.Span.End.Offset != len(source) || firstFile.Span.End.Line != 5 || firstFile.Span.End.Column != 1 {
		t.Fatalf("recovered file did not reach EOF: %#v", firstFile.Span)
	}
}
