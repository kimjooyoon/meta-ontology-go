package lsp

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

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
