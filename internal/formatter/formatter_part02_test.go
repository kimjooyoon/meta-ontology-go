package formatter

import (
	"testing"
)

type fixtureDeclaration struct {
	kind   DeclarationKind
	name   string
	id     string
	inputs []string
	output string
}
type fixtureAdapter struct{}

func (fixtureAdapter) Adapt(value any) (*Document, Diagnostics) {
	ast, ok := value.(*fixtureAST)
	if !ok || ast == nil {
		return nil, Diagnostics{{Severity: SeverityError, Code: CodeInvalidDocument, Message: "fixture AST has the wrong type"}}
	}
	document := Document{Package: ast.packageName, Namespace: ast.namespace}
	for _, declaration := range ast.declarations {
		document.Declarations = append(document.Declarations, Declaration{
			Kind: declaration.kind, Name: declaration.name, ID: declaration.id,
			Inputs: append([]string(nil), declaration.inputs...), Output: declaration.output,
		})
	}
	return &document, nil
}
func mustParseFixture(t *testing.T, source string) *fixtureAST {
	t.Helper()
	ast, diagnostics := parseFixture(source)
	if len(diagnostics) > 0 {
		t.Fatalf("fixture parse failed: %v", diagnostics)
	}
	return ast
}
func mustAdaptFixture(t *testing.T, ast *fixtureAST) *Document {
	t.Helper()
	document, diagnostics := (fixtureAdapter{}).Adapt(ast)
	if len(diagnostics) > 0 {
		t.Fatalf("fixture adaptation failed: %v", diagnostics)
	}
	return document
}
func billingDocument() Document {
	return Document{
		Package: "billing", Namespace: "billing",
		Declarations: []Declaration{
			{Kind: EntityDeclaration, Name: "Order", ID: "billing://entity/order"},
			{Kind: EntityDeclaration, Name: "Payment", ID: "billing://entity/payment"},
			{Kind: ActivityDeclaration, Name: "PayOrder", Inputs: []string{"Order"}, Output: "Payment"},
		},
	}
}
