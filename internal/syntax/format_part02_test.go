package syntax

import (
	"testing"
)

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
