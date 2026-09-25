package syntax

import (
	"strings"
	"testing"
)

func TestInvalidUnicodeEscapeDiagnosticUsesEscapeSpan(t *testing.T) {
	source := "package p\nnamespace n\nentity A id \"prefix\\u12xz\""
	_, diagnostics := LexFile("escape.gooo", source)
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagInvalidEscape {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	escapeOffset := strings.Index(source, `\u12xz`)
	if escapeOffset < 0 {
		t.Fatal("test source did not contain the invalid escape")
	}
	span := diagnostics[0].Span
	if span.Start.Offset != escapeOffset || span.End.Offset != escapeOffset+len(`\u12xz`) {
		t.Fatalf("invalid escape span = %#v, want byte range [%d,%d)", span, escapeOffset, escapeOffset+len(`\u12xz`))
	}
	if span.Start.Line != 3 || span.Start.Column >= span.End.Column {
		t.Fatalf("invalid escape position = %#v", span)
	}
}
