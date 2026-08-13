package lsp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"testing"
)

func TestServeMalformedPayloadPreservesNextFrameAndState(t *testing.T) {
	uri := "file:///malformed.gooo"
	var input, output bytes.Buffer
	writeFrameForTest(t, &input, []byte(`{"jsonrpc":"2.0","id":7,"method":"textDocument/didOpen","params":`))
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "package p\nnamespace n"},
	})
	writeRequest(t, &input, 8, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	server := NewServer()
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 || responseCode(t, messages[0]) != parseError {
		t.Fatalf("messages = %d, first error = %d", len(messages), responseCode(t, messages[0]))
	}
	assertDiagnostics(t, messages[1], uri, "")
	assertResultID(t, messages[2], 8)
	document, exists := server.documents[uri]
	if !exists || document.version != 1 || document.text != "package p\nnamespace n" {
		t.Fatalf("malformed payload changed document state: %#v", document)
	}
	var envelope struct{ ID json.RawMessage }
	decodeJSON(t, messages[0], &envelope)
	if string(envelope.ID) != "null" {
		t.Fatalf("parse error ID = %s, want null", envelope.ID)
	}
}

func TestServePartialPayloadRejectsWithoutOverlayMutation(t *testing.T) {
	uri := "file:///partial.gooo"
	source := "package p\nnamespace n"
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	partial := []byte(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///partial.gooo","version":2},"contentChanges":[{"text":"changed"}]}`)
	input.WriteString("Content-Length: " + strconv.Itoa(len(partial)+1) + "\r\n\r\n")
	_, _ = input.Write(partial)
	server := NewServer()
	err := server.Serve(&input, &output)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Serve() error = %v, want io.ErrUnexpectedEOF", err)
	}
	document, exists := server.documents[uri]
	if !exists || document.version != 1 || document.text != source {
		t.Fatalf("partial payload changed document state: %#v", document)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 1 {
		t.Fatalf("output messages = %d, want only open diagnostics", len(messages))
	}
	assertDiagnostics(t, messages[0], uri, "")
}
