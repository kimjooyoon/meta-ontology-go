package lsp

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

func runClosedFeature(t *testing.T, method string) {
	t.Helper()
	var calls atomic.Int32
	started, returned := make(chan struct{}), make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			return ParseResult{Symbols: []Symbol{{Name: "source", Detail: "ready"}}}, nil
		}
		close(started)
		<-ctx.Done()
		close(returned)
		return ParseResult{}, ctx.Err()
	})
	server := NewServer(parser)
	reader, writer := io.Pipe()
	output := newDiagnosticsBuffer()
	done := make(chan error, 1)
	go func() { done <- server.Serve(reader, output) }()
	uri := "file:///closed-" + method + ".gooo"
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "source"},
	})
	<-output.first
	writeRequest(t, writer, 7, method, featureRequestParams(method, uri))
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("feature parser did not start")
	}
	writeNotification(t, writer, "textDocument/didClose", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	<-output.second
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("closed feature parser did not observe cancellation")
	}
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	if responseIDPresent(t, output.Bytes(), 7) {
		t.Fatalf("closed feature emitted response: %q", output.Bytes())
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 {
		t.Fatalf("close lifecycle messages = %d, want open diagnostics, close diagnostics, shutdown", len(messages))
	}
	assertResultID(t, messages[2], 8)
	server.mu.RLock()
	_, exists := server.documents[uri]
	server.mu.RUnlock()
	if exists || calls.Load() != 2 {
		t.Fatalf("close mutated lifecycle unexpectedly: exists=%v parseCalls=%d", exists, calls.Load())
	}
}
