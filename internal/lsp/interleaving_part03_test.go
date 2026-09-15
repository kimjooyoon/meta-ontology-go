package lsp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

func TestServeContextReturnsWhenNonCloseableReaderBlocks(t *testing.T) {
	input := newNonCloseableReader()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- NewServer().ServeContext(ctx, input, &bytes.Buffer{}) }()
	<-input.started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ServeContext() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ServeContext() waited for a non-closeable reader after cancellation")
	}
	close(input.release)
	select {
	case <-input.finished:
	case <-time.After(time.Second):
		t.Fatal("non-closeable reader did not finish after release")
	}
}
func TestExitCancelsPendingRequest(t *testing.T) {
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
		"textDocument": map[string]any{"uri": "file:///exit.gooo", "version": 1, "text": "source"},
	})
	<-opened
	writeRequest(t, writer, 7, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": "file:///exit.gooo"}, "position": map[string]any{"character": 2},
	})
	<-started
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
