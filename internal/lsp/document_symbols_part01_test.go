package lsp

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestDocumentSymbolRequestUsesCanonicalParseResult(t *testing.T) {
	uri := "file:///symbols.gooo"
	symbols := []Symbol{
		{ID: "urn:later", Name: "Later", Kind: SymbolClass, Detail: "entity Later", Range: testRange(2, 1, 2, 12), SelectionRange: testRange(2, 8, 2, 13)},
		{ID: "urn:first", Name: "First", Kind: SymbolClass, Detail: "entity First", Range: testRange(1, 1, 1, 12), SelectionRange: testRange(1, 8, 1, 13)},
	}
	original := append([]Symbol(nil), symbols...)
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{Symbols: append([]Symbol(nil), symbols...)}
	})
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "initialize", nil)
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "canonical"},
	})
	writeRequest(t, &input, 2, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	writeRequest(t, &input, 3, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 4 {
		t.Fatalf("output messages = %d, want 4", len(messages))
	}
	assertInitialize(t, messages[0])
	var response struct {
		ID     int              `json:"id"`
		Result []DocumentSymbol `json:"result"`
	}
	decodeJSON(t, messages[2], &response)
	if response.ID != 2 || len(response.Result) != 2 {
		t.Fatalf("document symbols response = %#v", response)
	}
	if response.Result[0].ID != "" || !strings.Contains(response.Result[0].Detail, "urn:first") ||
		response.Result[1].ID != "" || !strings.Contains(response.Result[1].Detail, "urn:later") {
		t.Fatalf("document symbol wire identity projection = %#v", response.Result)
	}
	if response.Result[0].Name != "First" || !strings.HasPrefix(response.Result[0].Detail, "entity First") ||
		response.Result[0].Kind != SymbolClass || response.Result[0].SelectionRange.Start.Line != 1 {
		t.Fatalf("document symbol projection = %#v", response.Result[0])
	}
	if !reflect.DeepEqual(symbols, original) {
		t.Fatalf("parser symbols mutated: %#v", symbols)
	}
}
