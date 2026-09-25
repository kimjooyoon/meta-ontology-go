package lsp

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemanticTokensClosedDocumentSuppressesResponse(t *testing.T) {
	var calls atomic.Int32
	opened := make(chan struct{})
	started := make(chan struct{})
	returned := make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			close(opened)
			return ParseResult{Symbols: []Symbol{{Name: "source", Kind: SymbolClass, SelectionRange: Range{End: Position{Character: 6}}}}}, nil
		}
		close(started)
		<-ctx.Done()
		close(returned)
		return ParseResult{}, ctx.Err()
	})
	reader, writer := io.Pipe()
	output := newDiagnosticsBuffer()
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, output) }()
	uri := "file:///stale-tokens.gooo"
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "source"},
	})
	<-opened
	<-output.first
	writeRequest(t, writer, 7, "textDocument/semanticTokens/full", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("semantic token parser did not start")
	}
	writeNotification(t, writer, "textDocument/didClose", map[string]any{"textDocument": map[string]any{"uri": uri}})
	<-output.second
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("semantic token parser did not return")
	}
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	if responseIDPresent(t, output.Bytes(), 7) {
		t.Fatalf("closed semantic token request emitted a response: %q", output.Bytes())
	}
}
