package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSemanticTokensRequestUsesCanonicalASTViews(t *testing.T) {
	uri := "file:///tokens.gooo"
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols: []Symbol{
				{Name: "Order", Kind: SymbolClass, SelectionRange: Range{Start: Position{Character: 3}, End: Position{Character: 8}}},
				{Name: "Pay", Kind: SymbolFunction, SelectionRange: Range{Start: Position{Character: 9}, End: Position{Character: 12}}},
			},
			References: []Reference{{Name: "Order", Range: Range{Start: Position{Line: 1, Character: 2}, End: Position{Line: 1, Character: 7}}}},
		}
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "😀 Order Pay\n  Order"},
	})
	writeRequest(t, &input, 1, "textDocument/semanticTokens/full", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	writeRequest(t, &input, 2, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 {
		t.Fatalf("output messages = %d, want diagnostics, tokens, shutdown", len(messages))
	}
	var result SemanticTokens
	decodeJSON(t, responseResult(t, messages[1]), &result)
	want := []uint32{0, 3, 5, 0, 0, 0, 6, 3, 1, 0, 1, 2, 5, 2, 0}
	if !reflect.DeepEqual(result.Data, want) {
		t.Fatalf("semantic token data = %v, want %v", result.Data, want)
	}
}
func TestSemanticTokensMalformedAndEmptyResults(t *testing.T) {
	server := NewServer()
	cases := []string{"", "null", "{}", `{"textDocument":null}`, `{"textDocument":{}}`, `{"textDocument":[]}`}
	for _, raw := range cases {
		response, _, err := server.semanticTokensRequest(context.Background(), requestEnvelope{
			ID: json.RawMessage("1"), Params: json.RawMessage(raw),
		})
		if err != nil || responseCode(t, mustMarshalResponse(t, response)) != invalidParams {
			t.Fatalf("params %q: response=%#v err=%v", raw, response, err)
		}
	}
	uri := "file:///empty.gooo"
	server.documents[uri] = &document{result: ParseResult{}}
	response, _, err := server.semanticTokensRequest(context.Background(), requestEnvelope{
		ID: json.RawMessage("2"), Params: json.RawMessage(`{"textDocument":{"uri":"file:///empty.gooo"}}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	var result SemanticTokens
	decodeJSON(t, responseResult(t, mustMarshalResponse(t, response)), &result)
	if result.Data == nil || len(result.Data) != 0 {
		t.Fatalf("empty token data = %#v, want non-nil empty", result.Data)
	}
}
