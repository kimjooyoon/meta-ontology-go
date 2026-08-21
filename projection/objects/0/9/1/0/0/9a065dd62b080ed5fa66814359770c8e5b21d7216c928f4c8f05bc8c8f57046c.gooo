package syntax

import (
	"reflect"
	"testing"
)

func TestParseDiagnosticsTable(t *testing.T) {
	tests := []struct {
		name   string
		source string
		codes  []DiagnosticCode
	}{
		{
			name:   "missing headers",
			source: "",
			codes:  []DiagnosticCode{DiagExpectedPackage, DiagExpectedNamespace},
		},
		{
			name:   "missing entity id and string",
			source: "package p namespace n entity Thing",
			codes:  []DiagnosticCode{DiagExpectedID, DiagExpectedString},
		},
		{
			name:   "missing parameter comma",
			source: "package p namespace n activity A(One Two) -> Result",
			codes:  []DiagnosticCode{DiagExpectedComma},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Parse(test.source)
			got := make([]DiagnosticCode, len(diagnostics))
			for i, diagnostic := range diagnostics {
				got[i] = diagnostic.Code
			}
			if !reflect.DeepEqual(got, test.codes) {
				t.Fatalf("diagnostic codes = %v, want %v (%v)", got, test.codes, diagnostics)
			}
		})
	}
}
func TestParsePreservesDeclarationSpans(t *testing.T) {
	source := "package p\nnamespace n\nentity Order id \"urn:order\"\nactivity Pay(Order) -> Order\n"
	file, diagnostics := ParseFile("billing.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	activity := file.Declarations[1].(*ActivityDecl)
	if entity.Span.Filename != "billing.gooo" || entity.Span.Start.Line != 3 || entity.Span.End.Line != 3 {
		t.Fatalf("entity span = %#v", entity.Span)
	}
	if activity.Span.Start.Line != 4 || activity.Span.End.Line != 4 || activity.Parameters[0].Span.Start.Line != 4 {
		t.Fatalf("activity spans = %#v", activity)
	}
}
