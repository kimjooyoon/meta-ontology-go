package lsp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestServeLifecycleDiagnosticsAndFeatures(t *testing.T) {
	uri := "file:///billing.gooo"
	source := "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Order) -> Order\n"
	var input bytes.Buffer
	writeRequest(t, &input, 1, "initialize", map[string]any{})
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "languageId": "gooo", "version": 1, "text": source},
	})
	writeRequest(t, &input, 2, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 2, "character": 8},
	})
	writeRequest(t, &input, 3, "textDocument/completion", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 3, "character": 18},
	})
	writeRequest(t, &input, 4, "textDocument/definition", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 3, "character": 19},
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
		t.Fatalf("got %d output messages, want 7", len(messages))
	}
	assertInitializeResult(t, messages[0])
	assertDiagnostics(t, messages[1], uri, 0)
	assertHover(t, messages[2], "billing://entity/order")
	assertCompletion(t, messages[3], "Order")
	assertDefinition(t, messages[4], uri)
	assertDiagnostics(t, messages[5], uri, 1)
	assertResultID(t, messages[6], 5)
}

func TestRangeChangesUseUTF16Positions(t *testing.T) {
	source, err := applyChanges("😀x", []TextDocumentContentChangeEvent{{
		Range: &Range{Start: Position{Character: 2}, End: Position{Character: 3}}, Text: "y",
	}})
	if err != nil || source != "😀y" {
		t.Fatalf("applyChanges() = %q, error = %v", source, err)
	}
}

func TestParserSeamIsUsedForDiagnostics(t *testing.T) {
	called := false
	parser := ParserFunc(func(uri, source string) ParseResult {
		called = uri == "file:///test.gooo" && source == "source"
		return ParseResult{Diagnostics: []Diagnostic{{Message: "stub", Range: Range{}}}}
	})
	server := NewServer(parser)
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": "file:///test.gooo", "version": 1, "text": "source"},
	})
	writeNotification(t, &input, "exit", nil)
	_ = server.Serve(&input, &output)
	if !called {
		t.Fatal("parser was not called")
	}
}

func assertInitializeResult(t *testing.T, payload []byte) {
	t.Helper()
	var message struct {
		Result InitializeResult `json:"result"`
	}
	decode(t, payload, &message)
	if !message.Result.Capabilities.HoverProvider || !message.Result.Capabilities.DefinitionProvider {
		t.Fatalf("capabilities = %#v", message.Result.Capabilities)
	}
}

func assertDiagnostics(t *testing.T, payload []byte, uri string, want int) {
	t.Helper()
	var message struct {
		Method string                   `json:"method"`
		Params PublishDiagnosticsParams `json:"params"`
	}
	decode(t, payload, &message)
	if message.Method != "textDocument/publishDiagnostics" || message.Params.URI != uri || len(message.Params.Diagnostics) != want {
		t.Fatalf("diagnostics = %#v", message)
	}
}

func assertHover(t *testing.T, payload []byte, want string) {
	t.Helper()
	var message struct {
		Result *Hover `json:"result"`
	}
	decode(t, payload, &message)
	if message.Result == nil || !strings.Contains(message.Result.Contents.Value, want) {
		t.Fatalf("hover = %#v", message.Result)
	}
}

func assertCompletion(t *testing.T, payload []byte, want string) {
	t.Helper()
	var message struct {
		Result CompletionList `json:"result"`
	}
	decode(t, payload, &message)
	for _, item := range message.Result.Items {
		if item.Label == want {
			return
		}
	}
	t.Fatalf("completion items = %#v", message.Result.Items)
}

func assertDefinition(t *testing.T, payload []byte, uri string) {
	t.Helper()
	var message struct {
		Result []Location `json:"result"`
	}
	decode(t, payload, &message)
	if len(message.Result) != 1 || message.Result[0].URI != uri {
		t.Fatalf("definition = %#v", message.Result)
	}
}

func assertResultID(t *testing.T, payload []byte, want int) {
	t.Helper()
	var message struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	decode(t, payload, &message)
	if message.ID != want || string(message.Result) != "null" {
		t.Fatalf("response = %#v", message)
	}
}

func decode(t *testing.T, payload []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("decode %s: %v", payload, err)
	}
}
