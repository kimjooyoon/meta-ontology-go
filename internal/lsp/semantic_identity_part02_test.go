package lsp

import (
	"bytes"
	"strings"
	"testing"
)

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
