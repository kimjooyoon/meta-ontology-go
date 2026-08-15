package lsp

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

const billingIdentitySource = `package billing
namespace billing

entity Order id "billing://entity/order"
entity PaymentMethod id "billing://entity/payment-method"
entity Payment id "billing://entity/payment"

activity PayOrder(Order, PaymentMethod) -> Payment
`

func TestBillingTranscriptUsesOneLoweredIdentityModel(t *testing.T) {
	uri := "billing://open/main.gooo"
	input := billingTranscriptInput(t, uri)
	var output bytes.Buffer
	server := NewServer()
	if err := server.Serve(&input, &output); err != nil {
		t.Fatal(err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 9 {
		t.Fatalf("transcript messages = %d, want 9", len(messages))
	}
	assertDiagnostics(t, messages[1], uri, "")
	document := server.documents[uri]
	wantIDs := billingWantIDs()
	assertBillingDocument(t, document, wantIDs)
	assertBillingDocumentSymbols(t, messages[2], wantIDs)
	assertBillingNavigation(t, messages, document, wantIDs)
	assertBillingSemanticTokens(t, messages[7], document, wantIDs)
	if document.text != billingIdentitySource || document.version != 1 {
		t.Fatal("transcript changed the open overlay")
	}
}

func billingTranscriptInput(t *testing.T, uri string) bytes.Buffer {
	t.Helper()
	var input bytes.Buffer
	writeRequest(t, &input, 1, "initialize", nil)
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": billingIdentitySource},
	})
	writeRequest(t, &input, 2, "textDocument/documentSymbol", map[string]any{"textDocument": map[string]any{"uri": uri}})
	writeRequest(t, &input, 3, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 7, "character": 10},
	})
	writeRequest(t, &input, 4, "textDocument/completion", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 7, "character": 18},
	})
	writeRequest(t, &input, 5, "textDocument/definition", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 7, "character": 19},
	})
	writeRequest(t, &input, 6, "textDocument/references", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 3, "character": 7},
		"context": map[string]any{"includeDeclaration": true},
	})
	writeRequest(t, &input, 7, "textDocument/semanticTokens/full", map[string]any{"textDocument": map[string]any{"uri": uri}})
	writeRequest(t, &input, 8, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	return input
}

func billingWantIDs() map[string]string {
	return map[string]string{
		"Order":         "billing://entity/order",
		"PaymentMethod": "billing://entity/payment-method",
		"Payment":       "billing://entity/payment",
		"PayOrder":      "billing://activity/pay-order",
	}
}

func assertBillingDocument(t *testing.T, document *document, wantIDs map[string]string) {
	t.Helper()
	if document == nil || !document.result.semanticChecked || !document.result.semanticValid {
		t.Fatalf("document semantic state = %#v", document)
	}
	for name, wantID := range wantIDs {
		symbol, ok := symbolNamed(document.result.Symbols, name)
		if !ok || symbol.ID != wantID {
			t.Fatalf("%s identity = %#v, want %q", name, symbol, wantID)
		}
	}
}

func assertBillingDocumentSymbols(t *testing.T, message []byte, wantIDs map[string]string) {
	t.Helper()
	var response struct {
		Result []DocumentSymbol `json:"result"`
	}
	decodeJSON(t, message, &response)
	if bytes.Contains(responseResult(t, message), []byte(`"id"`)) {
		t.Fatalf("document symbols used a non-standard id field: %v", responseResult(t, message))
	}
	for name, wantID := range wantIDs {
		found := false
		for _, symbol := range response.Result {
			if symbol.Name == name {
				found = strings.Contains(symbol.Detail, wantID)
			}
		}
		if !found {
			t.Fatalf("document symbol %s did not expose %s: %#v", name, wantID, response.Result)
		}
	}
	if !hasDocumentSymbol(response.Result, "billing", SymbolNamespace) || !hasDocumentSymbol(response.Result, "billing", SymbolPackage) {
		t.Fatalf("package/namespace symbols = %#v", response.Result)
	}
}

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

func TestLoweredIdentityRenameAndNamespaceLocality(t *testing.T) {
	first := parseIdentityResult(t, `package billing
namespace billing
entity Order id "billing://entity/order"
`)
	second := parseIdentityResult(t, `package billing
namespace billing
entity Purchase id "billing://entity/order"
`)
	if first.Symbols[0].ID != second.Symbols[0].ID {
		t.Fatalf("explicit-ID rename changed identity: %q -> %q", first.Symbols[0].ID, second.Symbols[0].ID)
	}
	billing := parseIdentityResult(t, `package billing
namespace billing
entity Order id "billing://entity/order"
`)
	settlement := parseIdentityResult(t, `package settlement
namespace settlement
entity Order id "settlement://entity/order"
`)
	if billing.Symbols[0].ID == settlement.Symbols[0].ID {
		t.Fatalf("namespace-local declarations collapsed: %#v / %#v", billing.Symbols, settlement.Symbols)
	}
}

func TestMalformedOrUnknownLoweringHasDiagnosticsAndNoLinks(t *testing.T) {
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "unknown.gooo", `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Missing) -> Order
`)
	if err != nil || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "semantic.lowering" {
		t.Fatalf("unknown declaration diagnostics = %#v, error = %v", result.Diagnostics, err)
	}
	for _, symbol := range result.Symbols {
		if symbol.ID != "" || symbol.hasIdentity {
			t.Fatalf("unknown declaration invented symbol identity: %#v", symbol)
		}
	}
	for _, reference := range result.References {
		if reference.ID != "" {
			t.Fatalf("unknown declaration invented reference identity: %#v", reference)
		}
	}
}

func parseIdentityResult(t *testing.T, source string) ParseResult {
	t.Helper()
	result, err := (SyntaxParser{}).ParseContext(context.Background(), "identity.gooo", source)
	if err != nil || len(result.Diagnostics) != 0 || !result.semanticValid {
		t.Fatalf("identity parse = %#v, error = %v", result, err)
	}
	return result
}

func hasDocumentSymbol(symbols []DocumentSymbol, name string, kind SymbolKind) bool {
	for _, symbol := range symbols {
		if symbol.Name == name && symbol.Kind == kind {
			return true
		}
	}
	return false
}

func completionItem(items []CompletionItem, label string) (CompletionItem, bool) {
	for _, item := range items {
		if item.Label == label {
			return item, true
		}
	}
	return CompletionItem{}, false
}

func containsID(ids map[string]string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
