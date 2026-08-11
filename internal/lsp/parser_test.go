package lsp

import (
	"context"
	"reflect"
	"testing"
)

func TestSyntaxParserUsesCanonicalASTAndSpans(t *testing.T) {
	source := "package billing\r\nnamespace billing\r\nentity Order id \"billing://entity/order\"\r\nactivity Pay(Order) -> Order"
	first, err := (SyntaxParser{}).ParseContext(context.Background(), "billing.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	second, err := (SyntaxParser{}).ParseContext(context.Background(), "billing.gooo", source)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("parse is not deterministic: first=%#v second=%#v error=%v", first, second, err)
	}
	if first.File == nil || len(first.File.Declarations) != 2 || len(first.Diagnostics) != 0 {
		t.Fatalf("parse result = %#v", first)
	}
	entity := first.Symbols[0]
	if entity.Name != "Order" || entity.ID != "billing://entity/order" || entity.SelectionRange.Start != (Position{Line: 2, Character: 7}) {
		t.Fatalf("entity symbol = %#v", entity)
	}
	if len(first.References) != 2 || first.References[0].Range.Start != (Position{Line: 3, Character: 13}) {
		t.Fatalf("references = %#v", first.References)
	}
}

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
