package syntax

import (
	"reflect"
	"testing"
)

func TestDiagnosticsSortBySpanCanonicalEqualStartOrder(t *testing.T) {
	file, diagnostics := ParseFile("unicode.gooo", "package 도메인\r\nnamespace 도메인\r\n")
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	start := file.Span.End
	input := Diagnostics{
		{Code: "z-code", Message: "z", Severity: SeverityWarning, Span: eofSpan("z.gooo", start, 3)},
		{Code: "a-code", Message: "a", Severity: SeverityError, Span: eofSpan("z.gooo", start, 2)},
		{Code: "a-code", Message: "b", Severity: SeverityError, Span: eofSpan("z.gooo", start, 2)},
		{Code: "a-code", Message: "a", Severity: SeverityError, Span: eofSpan("a.gooo", start, 2)},
		{Code: "a-code", Message: "a", Severity: SeverityError, Span: eofSpan("z.gooo", start, 2)},
	}
	want := []string{"a.gooo:a", "z.gooo:a", "z.gooo:a", "z.gooo:b", "z.gooo:z"}
	got := diagnosticLabels(input.SortBySpan())
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical order = %v, want %v", got, want)
	}
	if start.Line != 3 || start.Column != 1 || start.Offset == 0 {
		t.Fatalf("test did not use Unicode CRLF EOF span: %#v", start)
	}
}
func TestDiagnosticsSortBySpanIsPermutationInvariant(t *testing.T) {
	base := Diagnostics{
		{Code: "c", Message: "end-3", Span: testDiagnosticSpan("same.gooo", 8, 3)},
		{Code: "a", Message: "end-1", Span: testDiagnosticSpan("same.gooo", 8, 1)},
		{Code: "b", Message: "warning", Severity: SeverityWarning, Span: testDiagnosticSpan("same.gooo", 8, 1)},
		{Code: "a", Message: "end-1-later", Span: testDiagnosticSpan("same.gooo", 8, 1)},
	}
	want := diagnosticLabels(base.SortBySpan())
	for _, permutation := range diagnosticPermutations(base) {
		if got := diagnosticLabels(permutation.SortBySpan()); !reflect.DeepEqual(got, want) {
			t.Fatalf("permutation order = %v, want %v", got, want)
		}
	}
}
func TestDiagnosticsSortBySpanReplaysWithoutMutation(t *testing.T) {
	source := "package p\nnamespace n\n"
	file, diagnostics := ParseFile("replay.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	input := Diagnostics{
		{Code: "eof", Message: "eof", Span: Span{Filename: "replay.gooo", Start: file.Span.End, End: file.Span.End}},
		{Code: "start", Message: "start", Span: file.Package.NameSpan},
	}
	original := append(Diagnostics(nil), input...)
	for range 3 {
		if got := diagnosticLabels(input.SortBySpan()); len(got) != 2 {
			t.Fatalf("unexpected replay result: %v", got)
		}
	}
	if !reflect.DeepEqual(input, original) || source != "package p\nnamespace n\n" {
		t.Fatal("sorting mutated diagnostic or source input")
	}
}
func eofSpan(filename string, start Position, length int) Span {
	return Span{Filename: filename, Start: start, End: Position{Offset: start.Offset + length, Line: start.Line, Column: start.Column + length}}
}
func testDiagnosticSpan(filename string, start, length int) Span {
	return eofSpan(filename, Position{Offset: start, Line: 2, Column: 4}, length)
}
