package formatter

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestFormatSyntaxBillingFixtureIsDeterministicAndSourceOrdered(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "examples", "billing", "main.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile("examples/billing/main.gooo", string(source))
	if len(diagnostics) != 0 {
		t.Fatalf("billing fixture diagnostics: %v", diagnostics)
	}

	first := FormatSyntax(file)
	if first.HasErrors() {
		t.Fatalf("billing formatter diagnostics: %v", first.Diagnostics)
	}
	if !strings.Contains(first.Source, "entity Order id \"billing://entity/order\"\nentity PaymentMethod") {
		t.Fatalf("formatter did not preserve entity source order:\n%s", first.Source)
	}
	if strings.Index(first.Source, "entity Order") > strings.Index(first.Source, "activity PayOrder") {
		t.Fatalf("formatter reordered declarations:\n%s", first.Source)
	}

	secondFile, secondDiagnostics := syntax.ParseFile("formatted.gooo", first.Source)
	if len(secondDiagnostics) != 0 {
		t.Fatalf("formatted billing diagnostics: %v\n%s", secondDiagnostics, first.Source)
	}
	second := FormatSyntax(secondFile)
	if second.HasErrors() || second.Source != first.Source {
		t.Fatalf("formatter is not idempotent:\nfirst=%q\nsecond=%q\ndiag=%v", first.Source, second.Source, second.Diagnostics)
	}
}

func TestFormatSyntaxPreservesDecodedLiteralValues(t *testing.T) {
	source := `package p
namespace n
entity Value id "urn:line\nquote\"slash\\nul\u0000"
`
	file, diagnostics := syntax.ParseFile("literal.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("literal source diagnostics: %v", diagnostics)
	}
	wantID := file.Declarations[0].(*syntax.EntityDecl).ID

	formatted := FormatSyntax(file)
	if formatted.HasErrors() || !strings.Contains(formatted.Source, `\u0000`) {
		t.Fatalf("literal formatting result: %#v", formatted)
	}
	roundTrip, roundTripDiagnostics := syntax.ParseFile("literal-formatted.gooo", formatted.Source)
	if len(roundTripDiagnostics) != 0 {
		t.Fatalf("formatted literal diagnostics: %v\n%s", roundTripDiagnostics, formatted.Source)
	}
	if got := roundTrip.Declarations[0].(*syntax.EntityDecl).ID; got != wantID {
		t.Fatalf("decoded literal changed: got %q, want %q", got, wantID)
	}
}

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
