package syntax

import (
	"strconv"
	"strings"
	"testing"
)

func TestLexRawStringPreservesTextAndSpans(t *testing.T) {
	value := "한글\\n\nfirst\r\nsecond\rthird"
	raw := "`" + value + "`"
	tokens, diagnostics := LexFile("raw.gooo", raw+" next")
	if len(diagnostics) != 0 || len(tokens) != 3 {
		t.Fatalf("raw literal did not produce one string: %v, %v", tokens, diagnostics)
	}
	got := tokens[0]
	if got.Kind != TokenString || got.Value != value || got.Text != raw ||
		got.Span.Start.Offset != 0 || got.Span.End.Offset != len(raw) ||
		got.Span.End.Line != 4 || got.Span.End.Column != 7 || tokens[1].Span.Start.Offset != len(raw)+1 {
		t.Fatalf("raw contents or source span changed: %#v", tokens)
	}
}

func TestLexRawStringKeepsEscapesLiteralAndReportsMalformedInput(t *testing.T) {
	for _, tc := range []struct {
		name, source, value string
		code                DiagnosticCode
	}{
		{"literal escapes", "`\\uD800 \\n \\\\ \"quoted\"`", "\\uD800 \\n \\\\ \"quoted\"", ""},
		{"unterminated", "`body\nnext", "body\nnext", DiagUnterminatedString},
		{"malformed UTF8", "`a" + string([]byte{0xff}) + "b`", "a\uFFFDb", DiagInvalidUTF8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tokens, diagnostics := LexFile("raw.gooo", tc.source)
			if tokens[0].Kind != TokenString || tokens[0].Value != tc.value || tokens[0].Text != tc.source {
				t.Fatalf("raw recovery changed literal bytes: %#v", tokens)
			}
			if tc.code == "" && len(diagnostics) != 0 || tc.code != "" &&
				(len(diagnostics) != 1 || diagnostics[0].Code != tc.code) {
				t.Fatalf("unexpected raw diagnostics: %v", diagnostics)
			}
			if tc.code == DiagInvalidUTF8 &&
				(diagnostics[0].Span.Start.Offset != 2 || diagnostics[0].Span.End.Offset != 3) {
				t.Fatalf("invalid byte lost its exact span: %v", diagnostics)
			}
		})
	}
}

func TestRawStringFormattingPreservesDecodedBodyAtFixedPoint(t *testing.T) {
	body := "\nlet message = \"한글\\n\"\r\nreturn message\n"
	source := "package raw\nnamespace raw\nentity Text id `raw://entity/text`\n" +
		"activity Message(Text) -> Text computes `" + body + "`"
	file, diagnostics := Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	first := file.Declarations[1].(*ActivityDecl)
	formatted, err := Format(file)
	if err != nil || !strings.Contains(formatted, strconv.Quote(body)) {
		t.Fatal("canonical formatting lost raw contents", err, formatted)
	}
	again, diagnostics := Parse(formatted)
	if diagnostics.HasErrors() || again.Declarations[1].(*ActivityDecl).ValueProgram != first.ValueProgram {
		t.Fatal("raw body changed across parse/format", diagnostics)
	}
	next, err := Format(again)
	if err != nil || formatted != next || again.Declarations[0].(*EntityDecl).ID != "raw://entity/text" {
		t.Fatal("raw parse/format is not a fixed point", err)
	}
}
