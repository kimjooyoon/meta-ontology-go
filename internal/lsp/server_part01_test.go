package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
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
func TestRefreshReusesExactDocumentAndInvalidatesChangedSource(t *testing.T) {
	uri := "file:///cache.gooo"
	source := "package p\nnamespace n\n"
	changed := "package p\nnamespace changed\n"
	calls := 0
	server := NewServer(ParserFunc(func(string, string) ParseResult {
		calls++
		return ParseResult{}
	}))
	params, err := json.Marshal(DidOpenTextDocumentParams{TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: source}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.didOpen(context.Background(), requestEnvelope{Params: params}); err != nil {
		t.Fatal(err)
	}
	calls = 0
	if err := server.refresh(context.Background(), uri); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("initial refresh parse calls = %d, want 1", calls)
	}
	calls = 0
	if err := server.refresh(context.Background(), uri); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("unchanged repeated refresh parse calls = %d, want 0", calls)
	}
	server.mu.Lock()
	server.documents[uri].text = changed
	server.documents[uri].version = 2
	server.mu.Unlock()
	if err := server.refresh(context.Background(), uri); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("changed refresh parse calls = %d, want 1", calls)
	}
	if err := server.refresh(context.Background(), uri); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("repeated changed refresh parse calls = %d, want 1", calls)
	}
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

func BenchmarkRefreshExactInput(b *testing.B) {
	uri := "file:///billing.gooo"
	sourceBytes, err := os.ReadFile("../../examples/billing/main.gooo")
	if err != nil {
		b.Fatal(err)
	}
	source := string(sourceBytes)
	calls := 0
	server := NewServer(ParserFunc(func(string, string) ParseResult {
		calls++
		return ParseResult{}
	}))
	params, err := json.Marshal(DidOpenTextDocumentParams{TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: source}})
	if err != nil {
		b.Fatal(err)
	}
	if _, _, err := server.didOpen(context.Background(), requestEnvelope{Params: params}); err != nil {
		b.Fatal(err)
	}
	if err := server.refresh(context.Background(), uri); err != nil {
		b.Fatal(err)
	}
	calls = 0
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := server.refresh(context.Background(), uri); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(calls)/float64(b.N), "parse-calls/op")
}
