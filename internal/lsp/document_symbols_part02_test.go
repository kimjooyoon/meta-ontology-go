package lsp

import (
	"bytes"
	"reflect"
	"testing"
)

func TestCanonicalDocumentSymbolsReplayAndNoMutation(t *testing.T) {
	symbols := []Symbol{
		{ID: "z", Name: "same", Detail: "z", Kind: SymbolClass, Range: testRange(1, 1, 1, 8), SelectionRange: testRange(1, 2, 1, 6)},
		{ID: "a", Name: "same", Detail: "a", Kind: SymbolClass, Range: testRange(1, 1, 1, 8), SelectionRange: testRange(1, 2, 1, 6)},
		{ID: "early", Name: "early", Kind: SymbolFunction, Range: testRange(0, 1, 0, 8), SelectionRange: testRange(0, 2, 0, 7)},
	}
	original := append([]Symbol(nil), symbols...)
	want := canonicalDocumentSymbols(symbols)
	for replay := 0; replay < 64; replay++ {
		got := canonicalDocumentSymbols(rotateSymbols(symbols, replay))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("replay %d differs:\n got %#v\nwant %#v", replay, got, want)
		}
	}
	if !reflect.DeepEqual(symbols, original) {
		t.Fatalf("canonical projection mutated input: %#v", symbols)
	}
	if want[0].ID != "early" || want[1].ID != "a" || want[2].ID != "z" {
		t.Fatalf("canonical order = %#v", want)
	}
}
func TestDocumentSymbolInvalidAndMissingDocuments(t *testing.T) {
	uri := "file:///invalid.gooo"
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{},
	})
	writeRequest(t, &input, 2, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": "file:///missing.gooo"},
	})
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "package p\nnamespace"},
	})
	writeRequest(t, &input, 3, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	writeRequest(t, &input, 4, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer().Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 5 {
		t.Fatalf("output messages = %d, want 5", len(messages))
	}
	if responseCode(t, messages[0]) != invalidParams {
		t.Fatalf("invalid params response code = %d", responseCode(t, messages[0]))
	}
	assertRawResult(t, messages[1], 2, "null")
	var response struct {
		ID     int              `json:"id"`
		Result []DocumentSymbol `json:"result"`
	}
	decodeJSON(t, messages[3], &response)
	if response.ID != 3 || len(response.Result) != 1 || response.Result[0].Name != "p" {
		t.Fatalf("malformed document symbols = %#v", response)
	}
}
