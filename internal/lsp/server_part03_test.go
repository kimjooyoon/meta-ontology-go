package lsp

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"sync/atomic"
	"testing"
)

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
