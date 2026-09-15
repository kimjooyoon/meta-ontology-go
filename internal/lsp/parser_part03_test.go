package lsp

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

func TestSyntaxDiagnosticsPreserveEOFAndCRLFSpans(t *testing.T) {
	source := "package billing\r\nnamespace billing\r\nentity Order id \"unterminated"
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "billing.gooo", source)
	if err != nil || len(result.Diagnostics) == 0 {
		t.Fatalf("diagnostics = %#v, error = %v", result.Diagnostics, err)
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Code != "lex.unterminated-string" || diagnostic.Source != "gooo" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if diagnostic.Range.Start != (Position{Line: 2, Character: 16}) || diagnostic.Range.End.Character <= 16 {
		t.Fatalf("diagnostic range = %#v", diagnostic.Range)
	}
}
func TestSyntaxDiagnosticsHandleInvalidUTF8Boundary(t *testing.T) {
	bytes := []byte("package p\nnamespace n\nentity A id \"x\" ")
	bytes = append(bytes, 0xff)
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "invalid.gooo", string(bytes))
	if err != nil || len(result.Diagnostics) == 0 {
		t.Fatalf("diagnostics = %#v, error = %v", result.Diagnostics, err)
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Code != "lex.unexpected-character" || diagnostic.Range.Start.Character != 16 {
		t.Fatalf("invalid byte diagnostic = %#v", diagnostic)
	}
}
func TestSyntaxEntityIdentitySpanHandlesInvalidUTF8WithoutMutation(t *testing.T) {
	bytes := []byte("package p\r\nnamespace n\r\nentity Order id \"urn:order\"\r\n")
	bytes = append(bytes, 0xff)
	source := string(bytes)
	file, diagnostics := syntax.ParseFile("invalid-identity.gooo", source)
	originalSource := source
	originalDiagnostics := append(syntax.Diagnostics(nil), diagnostics...)
	result, err := adaptSyntaxResult("invalid-identity.gooo", source, file, diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Symbols) != 1 {
		t.Fatalf("symbols = %#v", result.Symbols)
	}
	symbol := result.Symbols[0]
	if symbol.hasIdentity || symbol.ID != "" {
		t.Fatalf("malformed source invented identity = %#v", symbol)
	}
	if source != originalSource || !reflect.DeepEqual(diagnostics, originalDiagnostics) {
		t.Fatalf("identity mapping mutated source or diagnostics: source=%q diagnostics=%#v", source, diagnostics)
	}
}
