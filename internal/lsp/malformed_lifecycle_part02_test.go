package lsp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"testing"
)

func TestServePartialHeaderRejectsWithoutOverlayMutation(t *testing.T) {
	uri := "file:///partial-header.gooo"
	source := "package p\nnamespace n"
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	input.WriteString("Content-Length: 4\r\n")
	server := NewServer()
	err := server.Serve(&input, &output)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Serve() error = %v, want io.ErrUnexpectedEOF", err)
	}
	document, exists := server.documents[uri]
	if !exists || document.version != 1 || document.text != source {
		t.Fatalf("partial header changed document state: %#v", document)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 1 {
		t.Fatalf("output messages = %d, want only open diagnostics", len(messages))
	}
	assertDiagnostics(t, messages[0], uri, "")
}
func TestServeDuplicateHeaderRejectsWithoutOverlayMutation(t *testing.T) {
	uri := "file:///duplicate-header.gooo"
	source := "package p\nnamespace n"
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	payload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "textDocument/didChange",
		"params": map[string]any{
			"textDocument":   map[string]any{"uri": uri, "version": 2},
			"contentChanges": []map[string]any{{"text": "changed"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	length := strconv.Itoa(len(payload))
	input.WriteString("Content-Length: " + length + "\r\ncontent-length: " + length + "\r\n\r\n")
	_, _ = input.Write(payload)
	server := NewServer()
	err = server.Serve(&input, &output)
	if !errors.Is(err, ErrMalformedHeader) {
		t.Fatalf("Serve() error = %v, want ErrMalformedHeader", err)
	}
	document, exists := server.documents[uri]
	if !exists || document.version != 1 || document.text != source {
		t.Fatalf("duplicate header changed document state: %#v", document)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 1 {
		t.Fatalf("output messages = %d, want only open diagnostics", len(messages))
	}
	assertDiagnostics(t, messages[0], uri, "")
}
