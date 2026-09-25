package formatter

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

func TestFormatSyntaxRejectsCurrentEntityFieldsWithoutRewriting(t *testing.T) {
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	source := `package p
namespace n
entity Order id "urn:order" fields {
field Name id "urn:field/name" type string required one
}
`
	file, diagnostics := syntax.ParseWithEntityFieldsSupport(source, support)
	if len(diagnostics) != 0 || file == nil {
		t.Fatalf("entity fields setup failed: file=%#v diagnostics=%v", file, diagnostics)
	}
	original := file.Clone()
	result := FormatSyntax(file)
	if result.Source != "" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != CodeUnsupportedSyntax {
		t.Fatalf("unsupported entity fields result: %#v", result)
	}
	if !reflect.DeepEqual(file, original) {
		t.Fatal("unsupported entity fields formatting mutated the AST")
	}
}
func TestFormatSyntaxRejectsAliasConflictsWithoutRewriting(t *testing.T) {
	file, diagnostics := syntax.Parse(`package p namespace n activity Run() -> A`)
	if len(diagnostics) != 0 {
		t.Fatalf("alias setup diagnostics: %v", diagnostics)
	}
	activity := file.Declarations[0].(*syntax.ActivityDecl)
	activity.Output = "Other"
	original := file.Clone()
	result := FormatSyntax(file)
	if result.Source != "" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != CodeInvalidDocument {
		t.Fatalf("alias conflict result: %#v", result)
	}
	if !reflect.DeepEqual(file, original) {
		t.Fatal("alias conflict formatting mutated the AST")
	}
}
func TestFormatSyntaxRejectsInvalidLiteralAST(t *testing.T) {
	file, diagnostics := syntax.Parse(`package p namespace n entity A id "urn:a"`)
	if len(diagnostics) != 0 {
		t.Fatalf("invalid literal setup diagnostics: %v", diagnostics)
	}
	entity := file.Declarations[0].(*syntax.EntityDecl)
	entity.ID = string([]byte{'u', 'r', 'n', ':', 0xff})
	original := entity.ID
	result := FormatSyntax(file)
	if result.Source != "" || !result.HasErrors() || result.Diagnostics[0].Code != CodeInvalidDocument {
		t.Fatalf("invalid literal result: %#v", result)
	}
	if entity.ID != original {
		t.Fatal("invalid literal formatting mutated the AST")
	}
}
func TestSyntaxAdapterRejectsUnknownAST(t *testing.T) {
	result := FormatAST(struct{}{}, SyntaxAdapter{})
	if result.Source != "" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != CodeUnsupportedSyntax {
		t.Fatalf("unknown AST result: %#v", result)
	}
}
