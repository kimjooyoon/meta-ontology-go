package lsp

import (
	"bytes"
	"testing"
)

func TestServeRejectsDuplicateInitializeWithoutLosingDocumentState(t *testing.T) {
	uri := "file:///duplicate-initialize.gooo"
	source := "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\n"
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "initialize", map[string]any{})
	writeNotification(t, &input, "initialized", nil)
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeRequest(t, &input, 2, "initialize", map[string]any{})
	writeRequest(t, &input, 3, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 2, "character": 8},
	})
	writeRequest(t, &input, 4, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)

	server := NewServer()
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 5 {
		t.Fatalf("output messages = %d, want 5", len(messages))
	}
	assertInitialize(t, messages[0])
	assertDiagnostics(t, messages[1], uri, "")
	if got := responseCode(t, messages[2]); got != invalidRequest {
		t.Fatalf("duplicate initialize error code = %d, want %d", got, invalidRequest)
	}
	assertHover(t, messages[3], "entity Order")
	assertResultID(t, messages[4], 4)
	if document := server.documents[uri]; document == nil || document.text != source || document.version != 1 {
		t.Fatalf("document state after duplicate initialize = %#v", document)
	}
}
