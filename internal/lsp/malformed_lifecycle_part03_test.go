package lsp

import (
	"bytes"
	"errors"
	"testing"
)

func TestServeMalformedHeaderPreservesOverlay(t *testing.T) {
	uri := "file:///malformed-header.gooo"
	source := "package p\nnamespace n"
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	input.WriteString("Content-Length 4\r\n\r\ntest")
	server := NewServer()
	err := server.Serve(&input, &output)
	if !errors.Is(err, ErrMalformedHeader) {
		t.Fatalf("Serve() error = %v, want ErrMalformedHeader", err)
	}
	document, exists := server.documents[uri]
	if !exists || document.version != 1 || document.text != source {
		t.Fatalf("malformed header changed document state: %#v", document)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 1 {
		t.Fatalf("output messages = %d, want only open diagnostics", len(messages))
	}
	assertDiagnostics(t, messages[0], uri, "")
}
