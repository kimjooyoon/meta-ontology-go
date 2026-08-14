package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"sync/atomic"
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

func TestDocumentVersionAndRangeValidationPreserveOverlay(t *testing.T) {
	uri := "file:///version.gooo"
	source := "package p\nnamespace n\nentity A id \"urn:a\""
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeNotification(t, &input, "textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": 1},
		"contentChanges": []map[string]any{{"text": "stale"}},
	})
	writeNotification(t, &input, "textDocument/didChange", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 2},
		"contentChanges": []map[string]any{{
			"range": map[string]any{"start": map[string]any{"line": 0, "character": 0}, "end": map[string]any{"line": 0, "character": 99}},
			"text":  "bad",
		}},
	})
	writeRequest(t, &input, 1, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	server := NewServer()
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	document := server.documents[uri]
	if document == nil || document.version != 1 || document.text != source {
		t.Fatalf("overlay changed after invalid updates: %#v", document)
	}
	if messages := readFrames(t, output.Bytes()); len(messages) != 2 {
		t.Fatalf("output messages = %d, want open diagnostics and shutdown response", len(messages))
	}
}

func TestContextParserStopsOnCancellation(t *testing.T) {
	started := make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		close(started)
		<-ctx.Done()
		return ParseResult{}, ctx.Err()
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": "file:///cancel.gooo", "version": 1, "text": "source"},
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).ServeContext(ctx, &input, &output) }()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("ServeContext() error = %v, want context.Canceled", err)
	}
	if output.Len() != 0 {
		t.Fatalf("canceled parse produced output: %q", output.Bytes())
	}
}

func TestCancelRequestSuppressesInFlightResult(t *testing.T) {
	var calls atomic.Int32
	opened := make(chan struct{})
	started := make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			close(opened)
			return ParseResult{Symbols: []Symbol{{Name: "source", Detail: "ready"}}}, nil
		}
		close(started)
		<-ctx.Done()
		return ParseResult{}, ctx.Err()
	})
	reader, writer := io.Pipe()
	var output bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, &output) }()
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": "file:///cancel.gooo", "version": 1, "text": "source"},
	})
	<-opened
	writeRequest(t, writer, 7, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": "file:///cancel.gooo"}, "position": map[string]any{"character": 2},
	})
	<-started
	writeNotification(t, writer, "$/cancelRequest", map[string]any{"id": 7})
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	messages := readFrames(t, output.Bytes())
	if len(messages) != 2 {
		t.Fatalf("output messages = %d, want diagnostics and shutdown only", len(messages))
	}
	assertResultID(t, messages[1], 8)
}

func assertInitialize(t *testing.T, payload []byte) {
	t.Helper()
	var message struct {
		Result InitializeResult `json:"result"`
	}
	decodeJSON(t, payload, &message)
	if !message.Result.Capabilities.HoverProvider || !message.Result.Capabilities.DefinitionProvider ||
		!message.Result.Capabilities.DocumentSymbolProvider || !message.Result.Capabilities.ReferencesProvider ||
		message.Result.Capabilities.WorkspaceSymbolProvider == nil ||
		message.Result.Capabilities.WorkspaceSymbolProvider.Schema != WorkspaceSymbolProtocolSchema ||
		message.Result.Capabilities.SemanticTokensProvider == nil ||
		message.Result.Capabilities.SemanticTokensProvider.Schema != SemanticTokensProtocolSchema ||
		!message.Result.Capabilities.SemanticTokensProvider.Full {
		t.Fatalf("capabilities = %#v", message.Result.Capabilities)
	}
	wantTypes := canonicalSemanticTokenTypes
	gotTypes := message.Result.Capabilities.SemanticTokensProvider.Legend.TokenTypes
	if !reflect.DeepEqual(gotTypes, wantTypes) || message.Result.Capabilities.SemanticTokensProvider.Legend.TokenModifiers == nil {
		t.Fatalf("semantic token legend = %#v", message.Result.Capabilities.SemanticTokensProvider.Legend)
	}
	if message.Result.Capabilities.TextDocumentSync.Change != 2 {
		t.Fatalf("text document sync = %#v, want incremental change 2", message.Result.Capabilities.TextDocumentSync)
	}
}

func responseResult(t *testing.T, payload []byte) json.RawMessage {
	t.Helper()
	var response struct {
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &response)
	return response.Result
}

func assertDiagnostics(t *testing.T, payload []byte, uri, code string) {
	t.Helper()
	var message struct {
		Method string                   `json:"method"`
		Params PublishDiagnosticsParams `json:"params"`
	}
	decodeJSON(t, payload, &message)
	if message.Method != "textDocument/publishDiagnostics" || message.Params.URI != uri {
		t.Fatalf("diagnostic notification = %#v", message)
	}
	if code == "" && len(message.Params.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics = %#v", message.Params.Diagnostics)
	}
	if code != "" && (len(message.Params.Diagnostics) != 1 || message.Params.Diagnostics[0].Code != code) {
		t.Fatalf("diagnostics = %#v", message.Params.Diagnostics)
	}
}

func assertHover(t *testing.T, payload []byte, want string) {
	t.Helper()
	var message struct {
		Result *Hover `json:"result"`
	}
	decodeJSON(t, payload, &message)
	if message.Result == nil || !strings.Contains(message.Result.Contents.Value, want) {
		t.Fatalf("hover = %#v", message.Result)
	}
}

func assertCompletion(t *testing.T, payload []byte, want string) {
	t.Helper()
	var message struct {
		Result CompletionList `json:"result"`
	}
	decodeJSON(t, payload, &message)
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
	decodeJSON(t, payload, &message)
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
	decodeJSON(t, payload, &message)
	if message.ID != want || string(message.Result) != "null" {
		t.Fatalf("response = %#v", message)
	}
}
