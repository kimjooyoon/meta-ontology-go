package lsp

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

func TestInputEOFCancelsPendingRequest(t *testing.T) {
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
	output := newDiagnosticsBuffer()
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, output) }()
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": "file:///eof.gooo", "version": 1, "text": "source"},
	})
	<-opened
	<-output.first
	writeRequest(t, writer, 7, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": "file:///eof.gooo"}, "position": map[string]any{"character": 2},
	})
	<-started
	_ = writer.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Serve() did not cancel the pending request after input EOF")
	}
	if messages := readFrames(t, output.Bytes()); len(messages) != 1 {
		t.Fatalf("output messages = %d, want only open diagnostics", len(messages))
	}
}
