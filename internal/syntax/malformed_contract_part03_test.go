package syntax

import (
	"reflect"
	"testing"
)

func TestUnicodeEscapeRejectsSurrogateWithoutMutation(t *testing.T) {
	source := quotedIDSource("A", `\ud800`)
	original := source
	file, diagnostics := ParseFile("surrogate.gooo", source)
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagInvalidEscape {
		t.Fatalf("surrogate diagnostics = %#v", diagnostics)
	}
	if diagnostics[0].Span.Len() != len(`\ud800`) || file.Span.End.Offset != len(source) {
		t.Fatalf("surrogate spans = %#v, file = %#v", diagnostics[0].Span, file.Span)
	}
	formatted, formatDiagnostics, err := FormatSource("surrogate.gooo", source)
	if err == nil || formatted != "" || !reflect.DeepEqual(diagnostics, formatDiagnostics) {
		t.Fatalf("surrogate format result = %q, %#v, %v", formatted, formatDiagnostics, err)
	}
	secondFile, secondDiagnostics := ParseFile("surrogate.gooo", source)
	if !reflect.DeepEqual(file, secondFile) || !reflect.DeepEqual(diagnostics, secondDiagnostics) || source != original {
		t.Fatal("surrogate recovery was not deterministic or mutated input")
	}
}
func TestUnicodeScalarEscapeFormatsAndParsesToSameValue(t *testing.T) {
	source := quotedIDSource("A", `\u263a`)
	file, diagnostics := ParseFile("scalar.gooo", source)
	if len(diagnostics) != 0 || file.Declarations[0].(*EntityDecl).ID != "☺" {
		t.Fatalf("scalar parse = %#v, %#v", file, diagnostics)
	}
	formatted, formatDiagnostics, err := FormatSource("scalar.gooo", source)
	if err != nil || len(formatDiagnostics) != 0 {
		t.Fatalf("scalar format = %q, %#v, %v", formatted, formatDiagnostics, err)
	}
	roundTrip, roundTripDiagnostics := ParseFile("scalar-formatted.gooo", formatted)
	if len(roundTripDiagnostics) != 0 || roundTrip.Declarations[0].(*EntityDecl).ID != "☺" {
		t.Fatalf("scalar round-trip = %#v, %#v", roundTrip, roundTripDiagnostics)
	}
}
func TestFormatRejectsInvalidUTF8ASTIDWithoutMutation(t *testing.T) {
	file, diagnostics := Parse("package p namespace n entity A id \"urn:a\"")
	if len(diagnostics) != 0 {
		t.Fatalf("valid AST setup diagnostics = %#v", diagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	entity.ID = string([]byte{'u', 'r', 'n', ':', 0xff})
	original := entity.ID
	formatted, err := Format(file)
	if err == nil || formatted != "" {
		t.Fatalf("invalid UTF-8 AST format result = %q, %v", formatted, err)
	}
	if entity.ID != original {
		t.Fatal("formatting mutated invalid UTF-8 AST input")
	}
}
