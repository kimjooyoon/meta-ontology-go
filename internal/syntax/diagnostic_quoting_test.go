package syntax

import (
	"strconv"
	"strings"
	"testing"
)

func TestQuotedDiagnosticSourceRoundTripsWithoutControlCharacters(t *testing.T) {
	for _, source := range []string{"'", "\"", "line\nnext\r", "\x1b[31m", "한글 \\ 영어", "\u202e", string([]byte{0xff})} {
		quoted := quoteSource(source)
		decoded, err := strconv.Unquote(quoted)
		if err != nil || decoded != source {
			t.Fatal("diagnostic source changed", source, quoted, err)
		}
		if strings.ContainsAny(quoted, "\n\r\x1b") {
			t.Fatal("raw diagnostic control character", quoted)
		}
	}
}

func TestUnexpectedQuoteDiagnosticKeepsOriginalSpan(t *testing.T) {
	tokens, diagnostics := LexFile("quote.gooo", "'")
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagUnexpectedCharacter ||
		diagnostics[0].Message != `unexpected character "'"` ||
		diagnostics[0].Span.Start.Offset != 0 || diagnostics[0].Span.End.Offset != 1 ||
		diagnostics[0].Span.Filename != "quote.gooo" || tokens[len(tokens)-1].Kind != TokenEOF {
		t.Fatal("diagnostic quote or source location differs", diagnostics)
	}
}
