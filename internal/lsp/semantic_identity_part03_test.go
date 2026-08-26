package lsp

import (
	"bytes"
	"strings"
	"testing"
)

func assertBillingNavigation(t *testing.T, messages [][]byte, document *document, wantIDs map[string]string) {
	t.Helper()
	var hover struct {
		Result *Hover `json:"result"`
	}
	decodeJSON(t, messages[3], &hover)
	if hover.Result == nil || hover.Result.Contents.Value != "activity PayOrder" {
		t.Fatalf("activity hover = %#v", hover.Result)
	}
	if symbol, ok := symbolAtPosition(*document, Position{Line: 7, Character: 10}); !ok || symbol.ID != wantIDs["PayOrder"] {
		t.Fatalf("activity hover identity resolution = %#v/%v", symbol, ok)
	}
	var completion struct {
		Result CompletionList `json:"result"`
	}
	decodeJSON(t, messages[4], &completion)
	if item, ok := completionItem(completion.Result.Items, "PayOrder"); !ok || !strings.Contains(item.Documentation, wantIDs["PayOrder"]) {
		t.Fatalf("activity completion identity = %#v", item)
	}
	var definition struct {
		Result []Location `json:"result"`
	}
	decodeJSON(t, messages[5], &definition)
	order, _ := symbolNamed(document.result.Symbols, "Order")
	if len(definition.Result) != 1 || definition.Result[0].Range != order.SelectionRange {
		t.Fatalf("definition = %#v, want %v", definition.Result, order.SelectionRange)
	}
	var references struct {
		Result []Location `json:"result"`
	}
	decodeJSON(t, messages[6], &references)
	if len(references.Result) != 2 || references.Result[0].Range != order.SelectionRange {
		t.Fatalf("references = %#v, want declaration plus input", references.Result)
	}
}
func assertBillingSemanticTokens(t *testing.T, message []byte, document *document, wantIDs map[string]string) {
	t.Helper()
	var response struct {
		Result SemanticTokens `json:"result"`
	}
	decodeJSON(t, message, &response)
	if len(response.Result.Data)%5 != 0 || bytes.Contains(responseResult(t, message), []byte(`"semanticID"`)) {
		t.Fatalf("semantic tokens wire shape = %#v", response.Result)
	}
	for _, span := range semanticTokenSpansForDocument(*document) {
		if span.semanticID == "" || !containsID(wantIDs, span.semanticID) {
			t.Fatalf("token identity correspondence = %#v", span)
		}
	}
}
