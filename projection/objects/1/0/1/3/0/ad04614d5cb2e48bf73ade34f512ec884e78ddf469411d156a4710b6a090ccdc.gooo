package lsp

import (
	"bytes"
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
