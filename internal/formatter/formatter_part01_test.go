package formatter

import (
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
func TestFormatDocumentRejectsUnrepresentableActivityIdentity(t *testing.T) {
	document := billingDocument()
	document.Declarations[2].ID = "billing://activity/custom"
	result := Format(&document)
	if result.Source != "" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != CodeUnsupportedIdentity {
		t.Fatalf("unexpected activity identity result: %#v", result)
	}
}
func TestFormatDocumentNormalizesUnderscoreNamespaceForActivityIdentity(t *testing.T) {
	document := billingDocument()
	document.Package = "ci_time_causality"
	document.Namespace = "ci_time_causality"
	document.Declarations[2].Name = "BindOperationIdentity"
	result := Format(&document)
	if result.HasErrors() || !strings.Contains(result.Source, "activity BindOperationIdentity(Order) -> Payment") {
		t.Fatalf("underscore namespace was not formatted: %#v", result)
	}
}

type fixtureAST struct {
	packageName  string
	namespace    string
	declarations []fixtureDeclaration
}
