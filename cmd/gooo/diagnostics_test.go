package main

import (
	"bytes"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestFormatDiagnosticsCanonicalizesEqualStartPositions(t *testing.T) {
	span := func(endOffset, endColumn int) syntax.Span {
		return syntax.Span{
			Filename: "fixture.gooo",
			Start:    syntax.Position{Offset: 4, Line: 2, Column: 5},
			End:      syntax.Position{Offset: endOffset, Line: 2, Column: endColumn},
		}
	}
	input := syntax.Diagnostics{
		{Severity: syntax.SeverityWarning, Code: "diag.warning", Message: "warning", Span: span(8, 9)},
		{Severity: syntax.SeverityError, Code: "diag.z", Message: "z", Span: span(8, 9)},
		{Severity: syntax.SeverityError, Code: "diag.a", Message: "a", Span: span(8, 9)},
		{Severity: syntax.SeverityError, Code: "diag.short", Message: "short", Span: span(7, 8)},
	}
	reversed := append(syntax.Diagnostics(nil), input...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}

	first, err := formatDiagnostics(input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := formatDiagnostics(reversed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("equal-start diagnostics changed with insertion order:\nfirst=%ssecond=%s", first, second)
	}
	want := "fixture.gooo:2:5-2:8: error diag.short: short\n" +
		"fixture.gooo:2:5-2:9: error diag.a: a\n" +
		"fixture.gooo:2:5-2:9: error diag.z: z\n" +
		"fixture.gooo:2:5-2:9: warning diag.warning: warning\n"
	if string(first) != want {
		t.Fatalf("canonical diagnostics = %q, want %q", first, want)
	}
	if input[0].Code != "diag.warning" || input[1].Code != "diag.z" {
		t.Fatalf("canonicalization mutated input: %#v", input)
	}
}
