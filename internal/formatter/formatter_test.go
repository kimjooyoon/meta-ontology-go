package formatter

import (
	"strconv"
	"strings"
	"testing"
)

func TestFormatASTMissingASTReturnsDiagnostic(t *testing.T) {
	result := FormatAST(nil, fixtureAdapter{})
	if result.Source != "" || len(result.Diagnostics) != 1 {
		t.Fatalf("unexpected missing AST result: %#v", result)
	}
	if result.Diagnostics[0].Code != CodeMissingAST {
		t.Fatalf("unexpected diagnostic: %#v", result.Diagnostics)
	}
	var ast *fixtureAST
	result = FormatAST(ast, fixtureAdapter{})
	if result.Diagnostics[0].Code != CodeMissingAST {
		t.Fatalf("typed nil AST was not handled: %#v", result.Diagnostics)
	}
}

func TestFormatASTMissingAdapterReturnsDiagnostic(t *testing.T) {
	result := FormatAST(&fixtureAST{}, nil)
	if result.Source != "" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != CodeMissingAdapter {
		t.Fatalf("unexpected missing adapter result: %#v", result)
	}
}

func TestFormatDocumentUsesCanonicalOutput(t *testing.T) {
	document := billingDocument()
	result := Format(&document)
	want := "package billing\nnamespace billing\n\nentity Order id \"billing://entity/order\"\nentity Payment id \"billing://entity/payment\"\n\nactivity PayOrder(Order) -> Payment\n"
	if result.HasErrors() || result.Source != want {
		t.Fatalf("unexpected formatted result: %#v", result)
	}
}

func TestParseFormatParsePreservesSemanticFingerprint(t *testing.T) {
	source := "package billing\nnamespace billing\n\nentity Payment id \"billing://entity/payment\"\nentity Order id \"billing://entity/order\"\n\nactivity PayOrder(Order) -> Payment\n"
	firstAST := mustParseFixture(t, source)
	first := FormatAST(firstAST, fixtureAdapter{})
	if first.HasErrors() {
		t.Fatalf("format diagnostics: %v", first.Diagnostics)
	}
	secondAST := mustParseFixture(t, first.Source)
	firstDocument := mustAdaptFixture(t, firstAST)
	secondDocument := mustAdaptFixture(t, secondAST)
	if firstDocument.SemanticFingerprint() != secondDocument.SemanticFingerprint() {
		t.Fatalf("semantic meaning changed after format: before=%q after=%q", firstDocument.SemanticFingerprint(), secondDocument.SemanticFingerprint())
	}
}

func TestAdapterDiagnosticsPreventUnsafeOutput(t *testing.T) {
	adapter := ASTAdapterFunc(func(any) (*Document, Diagnostics) {
		return &Document{}, Diagnostics{{Severity: SeverityError, Code: CodeInvalidDocument, Message: "fixture error"}}
	})
	result := FormatAST(&fixtureAST{}, adapter)
	if result.Source != "" || !result.HasErrors() {
		t.Fatalf("adapter error was not propagated safely: %#v", result)
	}
}

type fixtureAST struct {
	packageName  string
	namespace    string
	declarations []fixtureDeclaration
}

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

func parseFixture(source string) (*fixtureAST, Diagnostics) {
	ast := &fixtureAST{}
	var diagnostics Diagnostics
	for lineNumber, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "package" {
			ast.packageName = fields[1]
			continue
		}
		if len(fields) == 2 && fields[0] == "namespace" {
			ast.namespace = fields[1]
			continue
		}
		if strings.HasPrefix(line, "entity ") {
			declaration, ok := parseFixtureEntity(fields)
			if !ok {
				diagnostics = appendFixtureError(diagnostics, lineNumber)
				continue
			}
			ast.declarations = append(ast.declarations, declaration)
			continue
		}
		if strings.HasPrefix(line, "activity ") {
			declaration, ok := parseFixtureActivity(line)
			if !ok {
				diagnostics = appendFixtureError(diagnostics, lineNumber)
				continue
			}
			ast.declarations = append(ast.declarations, declaration)
			continue
		}
		diagnostics = appendFixtureError(diagnostics, lineNumber)
	}
	return ast, diagnostics
}

func parseFixtureEntity(fields []string) (fixtureDeclaration, bool) {
	if len(fields) != 4 || fields[0] != "entity" || fields[2] != "id" {
		return fixtureDeclaration{}, false
	}
	decoded, err := strconv.Unquote(fields[3])
	if err != nil {
		return fixtureDeclaration{}, false
	}
	return fixtureDeclaration{kind: EntityDeclaration, name: fields[1], id: decoded}, decoded != ""
}

func parseFixtureActivity(line string) (fixtureDeclaration, bool) {
	parts := strings.Split(line, " -> ")
	if len(parts) != 2 {
		return fixtureDeclaration{}, false
	}
	left := strings.TrimPrefix(parts[0], "activity ")
	open := strings.Index(left, "(")
	close := strings.LastIndex(left, ")")
	if open <= 0 || close < open {
		return fixtureDeclaration{}, false
	}
	inputs := strings.TrimSpace(left[open+1 : close])
	var names []string
	if inputs != "" {
		for _, input := range strings.Split(inputs, ",") {
			names = append(names, strings.TrimSpace(input))
		}
	}
	return fixtureDeclaration{kind: ActivityDeclaration, name: left[:open], inputs: names, output: strings.TrimSpace(parts[1])}, true
}

func appendFixtureError(diagnostics Diagnostics, line int) Diagnostics {
	return append(diagnostics, Diagnostic{Severity: SeverityError, Code: CodeInvalidDocument, Message: "fixture parse error on line " + strconv.Itoa(line)})
}
