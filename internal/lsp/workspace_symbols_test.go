package lsp

import (
	"bytes"
	"encoding/json"
	"reflect"
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
		if got.Location.URI != expected.uri || got.Name != expected.name || got.ID != expected.id {
			t.Fatalf("workspace symbol %d = %#v, want %#v", index, got, expected)
		}
	}
	if response.Result[0].Detail != "entity Order" || response.Result[0].Kind != SymbolClass ||
		response.Result[0].Location.Range != testRange(0, 0, 0, 12) {
		t.Fatalf("workspace symbol projection = %#v", response.Result[0])
	}
	for uri, expected := range documents {
		if got := server.documents[uri].result.Symbols; !reflect.DeepEqual(got, expected) {
			t.Fatalf("workspace symbol request mutated %s: %#v", uri, got)
		}
	}
}

func TestWorkspaceSymbolsStrictParamsAndStableEmptyResults(t *testing.T) {
	server := NewServer()
	for _, raw := range []string{"", "null", "{}", `{"query":null}`, `{"query":7}`} {
		response, _, err := server.workspaceSymbolRequest(requestEnvelope{
			ID: json.RawMessage("1"), Params: json.RawMessage(raw),
		})
		if err != nil || response == nil || response.Error == nil || response.Error.Code != invalidParams {
			t.Fatalf("raw params %q response = %#v, error = %v", raw, response, err)
		}
	}
	response, _, err := server.workspaceSymbolRequest(requestEnvelope{
		ID: json.RawMessage("2"), Params: json.RawMessage(`{"query":"missing"}`),
	})
	if err != nil || response == nil || string(response.Result) != "[]" {
		t.Fatalf("empty workspace result = %#v, error = %v", response, err)
	}
}

func TestCanonicalWorkspaceSymbolsReplayAndNoMutation(t *testing.T) {
	symbols := []WorkspaceSymbol{
		{ID: "z", Name: "Same", Location: Location{URI: "file:///z.gooo", Range: testRange(0, 0, 0, 4)}},
		{ID: "a", Name: "Same", Location: Location{URI: "file:///a.gooo", Range: testRange(0, 0, 0, 4)}},
		{ID: "b", Name: "Same", Location: Location{URI: "file:///a.gooo", Range: testRange(1, 0, 1, 4)}},
	}
	original := append([]WorkspaceSymbol(nil), symbols...)
	want := canonicalWorkspaceSymbols(symbols)
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	for replay := 0; replay < 64; replay++ {
		gotJSON, err := json.Marshal(canonicalWorkspaceSymbols(rotateWorkspaceSymbols(symbols, replay)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(gotJSON, wantJSON) {
			t.Fatalf("replay %d = %s, want %s", replay, gotJSON, wantJSON)
		}
	}
	if !reflect.DeepEqual(symbols, original) {
		t.Fatalf("canonical workspace symbols mutated input: %#v", symbols)
	}
}

func rotateWorkspaceSymbols(symbols []WorkspaceSymbol, offset int) []WorkspaceSymbol {
	result := make([]WorkspaceSymbol, len(symbols))
	for index := range symbols {
		result[index] = symbols[(index+offset)%len(symbols)]
	}
	return result
}
