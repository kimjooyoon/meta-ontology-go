package syntax

import (
	"reflect"
	"testing"
)

func TestStage0SyntaxContractPreservesSemanticShape(t *testing.T) {
	source := "package billing\r\nnamespace billing\r\nentity Order id \"urn:order\"\r\nactivity Pay(Order) -> Order\r\n"
	file, diagnostics := ParseFile("billing.gooo", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if file.Span.Filename != "billing.gooo" || file.Span.End.Line != 5 || file.Span.End.Column != 1 {
		t.Fatalf("file span = %#v", file.Span)
	}
	if len(file.Decls) != len(file.Declarations) || len(file.Declarations) != 2 {
		t.Fatalf("declaration aliases diverged: %#v %#v", file.Decls, file.Declarations)
	}
	for index := range file.Decls {
		if file.Decls[index] != file.Declarations[index] {
			t.Fatalf("declaration alias at index %d does not preserve identity", index)
		}
	}
	entity, ok := file.Declarations[0].(*EntityDecl)
	if !ok || entity.Name != "Order" || entity.ID != "urn:order" || entity.Span.Start.Line != 3 {
		t.Fatalf("entity shape = %#v", file.Declarations[0])
	}
	activity, ok := file.Declarations[1].(*ActivityDecl)
	if !ok || activity.Name != "Pay" || activity.Output != activity.Result.Name {
		t.Fatalf("activity result aliases diverged: %#v", file.Declarations[1])
	}
	if !reflect.DeepEqual(activity.Inputs, activity.Parameters) || len(activity.Parameters) != 1 || activity.Parameters[0].Name != "Order" {
		t.Fatalf("activity input aliases diverged: %#v", activity)
	}
	if activity.Span.Start.Line != 4 || activity.Span.End.Line != 4 || activity.Result.Span.Start.Line != 4 {
		t.Fatalf("activity source spans = %#v", activity)
	}
}
func TestStage0SyntaxContractDiagnosticsAreDeterministic(t *testing.T) {
	source := "package p\nnamespace n\nactivity Pay(Order Missing) -> Order"
	firstFile, firstDiagnostics := ParseFile("invalid.gooo", source)
	secondFile, secondDiagnostics := ParseFile("invalid.gooo", source)
	if !reflect.DeepEqual(firstFile, secondFile) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("equivalent parser inputs produced different results")
	}
	if len(firstDiagnostics) != 1 || firstDiagnostics[0].Code != DiagExpectedComma {
		t.Fatalf("diagnostics = %#v, want one expected-comma diagnostic", firstDiagnostics)
	}
	activity, ok := firstFile.Declarations[0].(*ActivityDecl)
	if !ok || len(activity.Parameters) != 1 || activity.Result.Name != "Order" {
		t.Fatalf("recovered activity = %#v", firstFile.Declarations)
	}
}
func TestStage0SyntaxContractLexerIsIdempotentAndDefensive(t *testing.T) {
	lexer := NewLexer("entity Order id \"urn:order\"")
	firstTokens, firstDiagnostics := lexer.Lex()
	secondTokens, secondDiagnostics := lexer.Lex()
	if !reflect.DeepEqual(firstTokens, secondTokens) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("repeated lexing produced different results")
	}
	firstTokens[0].Text = "mutated"
	thirdTokens, _ := lexer.Lex()
	if thirdTokens[0].Text == "mutated" {
		t.Fatal("lexer exposed its cached token slice")
	}
	if thirdTokens[len(thirdTokens)-1].Kind != TokenEOF {
		t.Fatalf("tokens do not end in EOF: %#v", thirdTokens)
	}
}
