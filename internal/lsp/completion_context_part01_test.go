package lsp

import (
	"context"
	"testing"
)

func TestCompletionUsesIRBackedEntityCandidatesInActivityContext(t *testing.T) {
	uri := "file:///billing.gooo"
	source := "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nactivity Pay(Order) -> Order\n"
	parsed, err := (SyntaxParser{}).ParseContext(context.Background(), uri, source)
	if err != nil || !parsed.semanticValid {
		t.Fatalf("parse result = %#v, error = %v", parsed, err)
	}
	server := NewServer()
	server.documents[uri] = &document{text: source, result: parsed}
	position := Position{Line: 3, Character: len("activity Pay(")}
	items := server.completionAt(uri, position, true).Items
	if item, ok := completionItem(items, "Order"); !ok || item.Kind != int(SymbolClass) {
		t.Fatalf("entity completion = %#v", item)
	}
	if _, ok := completionItem(items, "Pay"); ok {
		t.Fatal("activity completion leaked the enclosing activity symbol")
	}
}

func TestCompletionNarrowsEntityFieldDeclarationCandidates(t *testing.T) {
	uri := "file:///fields.gooo"
	source := "package billing\nnamespace billing\nentity Order id \"billing://entity/order\" fields {\n  field "
	server := NewServer()
	server.documents[uri] = &document{
		text: source,
		result: ParseResult{Symbols: []Symbol{
			{Name: "Order", Kind: SymbolClass},
			{Name: "total", Kind: SymbolField},
			{Name: "Pay", Kind: SymbolFunction},
		}},
	}
	items := server.completionAt(uri, Position{Line: 3, Character: len("  field ")}, true).Items
	if item, ok := completionItem(items, "total"); !ok || item.Kind != int(SymbolField) {
		t.Fatalf("field completion = %#v", item)
	}
	if _, ok := completionItem(items, "Order"); ok {
		t.Fatal("entity symbol leaked into field declaration completion")
	}
	if _, ok := completionItem(items, "Pay"); ok {
		t.Fatal("activity symbol leaked into field declaration completion")
	}
}