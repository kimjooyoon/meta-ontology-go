package lsp

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestWorkspaceSymbolsMatchOpenDocumentsByNameAndID(t *testing.T) {
	documents := map[string][]Symbol{
		"file:///z.gooo": {
			{Name: "Order", ID: "id-z", Kind: SymbolClass, Range: testRange(2, 0, 2, 12), SelectionRange: testRange(2, 0, 2, 5)},
			{Name: "Order", ID: "id-a", Kind: SymbolClass, Range: testRange(1, 0, 1, 12), SelectionRange: testRange(1, 0, 1, 5)},
		},
		"file:///a.gooo": {
			{Name: "Order", ID: "id-b", Kind: SymbolClass, Detail: "entity Order", Range: testRange(0, 0, 0, 12), SelectionRange: testRange(0, 0, 0, 5)},
			{Name: "Other", ID: "order-id", Kind: SymbolFunction, Range: testRange(1, 0, 1, 12), SelectionRange: testRange(1, 0, 1, 5)},
		},
	}
	parser := ParserFunc(func(uri, _ string) ParseResult {
		return ParseResult{Symbols: append([]Symbol(nil), documents[uri]...)}
	})
	var input, output bytes.Buffer
	for _, uri := range []string{"file:///z.gooo", "file:///a.gooo"} {
		writeNotification(t, &input, "textDocument/didOpen", map[string]any{
			"textDocument": map[string]any{"uri": uri, "version": 1, "text": "open"},
		})
	}
	writeRequest(t, &input, 1, "workspace/symbol", map[string]any{"query": "OrDeR"})
	writeRequest(t, &input, 2, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	server := NewServer(parser)
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 4 {
		t.Fatalf("output messages = %d, want 4", len(messages))
	}
	var response struct {
		ID     int               `json:"id"`
		Result []WorkspaceSymbol `json:"result"`
	}
	decodeJSON(t, messages[2], &response)
	if response.ID != 1 || len(response.Result) != 4 {
		t.Fatalf("workspace symbols response = %#v", response)
	}
	want := []struct{ uri, name, id string }{
		{"file:///a.gooo", "Order", "id-b"},
		{"file:///a.gooo", "Other", "order-id"},
		{"file:///z.gooo", "Order", "id-a"},
		{"file:///z.gooo", "Order", "id-z"},
	}
	for index, expected := range want {
		got := response.Result[index]
		if got.Location.URI != expected.uri || got.Name != expected.name || got.ID != "" || !strings.Contains(got.Detail, expected.id) {
			t.Fatalf("workspace symbol %d = %#v, want %#v", index, got, expected)
		}
	}
	if !strings.HasPrefix(response.Result[0].Detail, "entity Order") || response.Result[0].Kind != SymbolClass ||
		response.Result[0].Location.Range != testRange(0, 0, 0, 12) {
		t.Fatalf("workspace symbol projection = %#v", response.Result[0])
	}
	for uri, expected := range documents {
		if got := server.documents[uri].result.Symbols; !reflect.DeepEqual(got, expected) {
			t.Fatalf("workspace symbol request mutated %s: %#v", uri, got)
		}
	}
}
