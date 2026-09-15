package syntax

import (
	"reflect"
	"testing"
)

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
