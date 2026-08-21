package lsp

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestServerLifecycleDiagnosticsAndFeatures(t *testing.T) {
	uri := "file:///billing.gooo"
	source := "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Order) -> Order\n"
	var input bytes.Buffer
	writeRequest(t, &input, 1, "initialize", map[string]any{})
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeRequest(t, &input, 2, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 2, "character": 8},
	})
	writeRequest(t, &input, 3, "textDocument/completion", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 3, "character": 14},
	})
	writeRequest(t, &input, 4, "textDocument/definition", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 3, "character": 14},
	})
	writeNotification(t, &input, "textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": 2},
		"contentChanges": []map[string]any{{"text": "package billing\nnamespace billing\nentity Order id \"unterminated"}},
	})
	writeRequest(t, &input, 5, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)

	var output bytes.Buffer
	if err := NewServer().Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 7 {
		t.Fatalf("output messages = %d, want 7", len(messages))
	}
	assertInitialize(t, messages[0])
	assertDiagnostics(t, messages[1], uri, "")
	assertHover(t, messages[2], "entity Order")
	assertCompletion(t, messages[3], "Order")
	assertDefinition(t, messages[4], uri)
	assertDiagnostics(t, messages[5], uri, "lex.unterminated-string")
	assertResultID(t, messages[6], 5)
}
func TestInitializeAdvertisesReadFeaturesAndDefersSourceMaps(t *testing.T) {
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "initialize", nil)
	writeRequest(t, &input, 2, "workspace/symbol", map[string]any{"query": "Order"})
	writeRequest(t, &input, 3, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer().Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 || string(responseResult(t, messages[1])) != "[]" {
		t.Fatalf("messages = %d, workspace result = %s", len(messages), responseResult(t, messages[1]))
	}
	var envelope struct {
		Result struct {
			Capabilities map[string]json.RawMessage `json:"capabilities"`
		} `json:"result"`
	}
	decodeJSON(t, messages[0], &envelope)
	for _, unsupported := range []string{"sourceMapProvider"} {
		if _, advertised := envelope.Result.Capabilities[unsupported]; advertised {
			t.Fatalf("unsupported capability %q was advertised", unsupported)
		}
	}
}
