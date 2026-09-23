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
