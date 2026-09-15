package formatter

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
