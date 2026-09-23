package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestCompletionUsesCanonicalSyntaxKeywords(t *testing.T) {
	items := serverCompletionLabels(NewServer().completion("missing://document"))
	for _, keyword := range syntax.CanonicalKeywordNames() {
		if !items[keyword] {
			t.Fatalf("completion omitted canonical keyword %q", keyword)
		}
	}
}

func TestCompletionFiltersByCursorPrefix(t *testing.T) {
	server := NewServer()
	uri := "file:///prefix.gooo"
	server.documents[uri] = &document{
		text: "Pay",
		result: ParseResult{Symbols: []Symbol{
			{Name: "PayOrder", Kind: SymbolFunction},
			{Name: "Order", Kind: SymbolClass},
		}},
	}
	items := server.completionAt(uri, Position{Line: 0, Character: 3}, true)
	if _, ok := completionItem(items.Items, "PayOrder"); !ok {
		t.Fatalf("prefix completion omitted matching symbol: %#v", items.Items)
	}
	if _, ok := completionItem(items.Items, "Order"); ok {
		t.Fatalf("prefix completion retained non-matching symbol: %#v", items.Items)
	}
}

func serverCompletionLabels(list *CompletionList) map[string]bool {
	labels := make(map[string]bool, len(list.Items))
	for _, item := range list.Items {
		labels[item.Label] = true
	}
	return labels
}
