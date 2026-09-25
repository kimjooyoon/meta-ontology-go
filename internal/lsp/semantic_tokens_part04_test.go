package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync/atomic"
	"testing"
)

func TestSemanticTokensCancelSuppressesResult(t *testing.T) {
	var calls atomic.Int32
	opened, started := make(chan struct{}), make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			close(opened)
			return ParseResult{}, nil
		}
		close(started)
		<-ctx.Done()
		return ParseResult{}, ctx.Err()
	})
	reader, writer := io.Pipe()
	var output bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, &output) }()
	uri := "file:///cancel-tokens.gooo"
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "source"},
	})
	<-opened
	writeRequest(t, writer, 7, "textDocument/semanticTokens/full", map[string]any{
		"textDocument": map[string]any{"uri": uri},
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
		t.Fatalf("output messages = %d, want diagnostics and shutdown", len(messages))
	}
	assertResultID(t, messages[1], 8)
}
func mustMarshalResponse(t *testing.T, response *responseEnvelope) []byte {
	t.Helper()
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
