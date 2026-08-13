package syntax

import "testing"

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

func TestFormatRejectsInvalidASTAliases(t *testing.T) {
	file, diagnostics := Parse("package p namespace n activity Run() -> A")
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	activity := file.Declarations[0].(*ActivityDecl)
	activity.Output = "Other"
	if _, err := Format(file); err == nil {
		t.Fatal("conflicting result aliases were accepted")
	}
}

func TestFormatRejectsConflictingFileDeclarationAliases(t *testing.T) {
	file, diagnostics := Parse("package p namespace n entity A id \"urn:a\"")
	if len(diagnostics) != 0 {
		t.Fatalf("valid AST setup diagnostics = %#v", diagnostics)
	}
	canonical := append([]Declaration(nil), file.Decls...)
	file.Declarations = append(append([]Declaration(nil), canonical...), canonical[0])
	formatted, err := Format(file)
	if err == nil || formatted != "" {
		t.Fatalf("conflicting file aliases format result = %q, %v", formatted, err)
	}
	if len(file.Decls) != len(canonical) || len(file.Declarations) != len(canonical)+1 {
		t.Fatal("formatting mutated declaration aliases")
	}
}

func TestFormatSourceReturnsDiagnosticsWithoutOutput(t *testing.T) {
	formatted, diagnostics, err := FormatSource("invalid.gooo", "package p namespace n entity Broken id \"")
	if err == nil || len(diagnostics) == 0 || formatted != "" {
		t.Fatalf("invalid source result = %q, %v, %v", formatted, diagnostics, err)
	}
}
