package bodycodegen

import (
	"context"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestSourceRenameEscapesKeepStringValues(t *testing.T) {
	for _, value := range []string{
		`\b`, `\f`, `\\b`, `C:\billing\fees`, `return "\\b"`,
		"\b\f\x00\x01\t\r\n", "가격 😀 <>&", "\\\b\\\f",
	} {
		literal := renameString(value, "")
		tokens, diagnostics := syntax.Lex(literal)
		if diagnostics.HasErrors() || len(tokens) != 2 || tokens[0].Value != value {
			t.Fatalf("quoted value changed: input=%q literal=%s tokens=%+v diagnostics=%v", value, literal, tokens, diagnostics)
		}
	}
	body := `let text = "\\b\\f"; _ = text; return Calculate(input)`
	source := strings.Replace(sourceRenameFixture,
		`computes "return input1 == Calculate(input0)"`, "computes "+renameString(body, ""), 1)
	result, err := RenameActivitySource(context.Background(), "escapes.gooo", []byte(source), "Calculate", "요금")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := ParseBodyFile("renamed.gooo", []byte(result.Source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	want := strings.Replace(body, "Calculate(input)", "요금(input)", 1)
	actual := file.Declarations[3].(*syntax.ActivityDecl).ValueProgram
	if actual != want {
		t.Fatalf("body changed beyond call name:\nactual=%q\nwant=%q", actual, want)
	}
}
