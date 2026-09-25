package syntax

import (
	"testing"
)

func TestFormatParseFixedPoint(t *testing.T) {
	sources := []string{
		"package billing\nnamespace billing\nentity Order id \"urn:order\"\nactivity Pay(Order) -> Order\n",
		"  package billing\r\nnamespace billing\r\n\r\nentity Demo id \"urn:a\\nquote\\\"\\\\b\"\r\nactivity Tick() -> Demo\r\n",
	}
	for _, source := range sources {
		first, diagnostics, err := FormatSource("fixture.gooo", source)
		if len(diagnostics) != 0 {
			t.Fatalf("source diagnostics = %v", diagnostics)
		}
		if err != nil {
			t.Fatal(err)
		}
		secondFile, secondDiagnostics := ParseFile("formatted.gooo", first)
		if len(secondDiagnostics) != 0 {
			t.Fatalf("formatted diagnostics = %v\n%s", secondDiagnostics, first)
		}
		second, err := Format(secondFile)
		if err != nil {
			t.Fatal(err)
		}
		if first != second {
			t.Fatalf("format is not a fixed point:\nfirst:\n%q\nsecond:\n%q", first, second)
		}
	}
}
func TestFormatPreservesDecodedEntityID(t *testing.T) {
	source := `package p
namespace n
entity A id "urn:line\nquote\"slash\\nul\u0000"`
	file, diagnostics := Parse(source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatal(err)
	}
	roundTripped, diagnostics := Parse(formatted)
	if len(diagnostics) != 0 {
		t.Fatalf("formatted diagnostics: %v\n%s", diagnostics, formatted)
	}
	first := file.Declarations[0].(*EntityDecl)
	second := roundTripped.Declarations[0].(*EntityDecl)
	if first.ID != second.ID || first.IDSpan.IsEmpty() || second.IDSpan.IsEmpty() {
		t.Fatalf("entity ID changed across formatting: %q != %q", first.ID, second.ID)
	}
}
func TestFormatCanonicalLayout(t *testing.T) {
	file, diagnostics := Parse("package p namespace n entity A id \"urn:a\" activity Run(A) -> A")
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatal(err)
	}
	want := "package p\nnamespace n\n\nentity A id \"urn:a\"\nactivity Run(A) -> A\n"
	if formatted != want {
		t.Fatalf("formatted source = %q, want %q", formatted, want)
	}
}
