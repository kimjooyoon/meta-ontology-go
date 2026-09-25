package lsp

import (
	"bytes"
	"testing"
)

func TestDogfoodDiagnosticsReplayAndUnsupportedNavigationFailClosed(t *testing.T) {
	uri := "file:///dogfood-diagnostics.gooo"
	source := "package p\r\nnamespace n\r\nentity Order id \"unterminated\r\n"
	first := serveDogfoodDiagnosticsSession(t, uri, source)
	second := serveDogfoodDiagnosticsSession(t, uri, source)
	if !bytes.Equal(first, second) {
		t.Fatalf("diagnostic replay differs:\nfirst=%s\nsecond=%s", first, second)
	}
	messages := readFrames(t, first)
	if len(messages) != 4 {
		t.Fatalf("diagnostic dogfood output messages = %d, want 4", len(messages))
	}
	assertDiagnostics(t, messages[1], uri, "lex.unterminated-string")
	if got := responseCode(t, messages[2]); got != methodNotFound {
		t.Fatalf("unsupported navigation error code = %d, want %d", got, methodNotFound)
	}
	assertResultID(t, messages[3], 3)
}
func serveDogfoodDiagnosticsSession(t *testing.T, uri, source string) []byte {
	t.Helper()
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "initialize", nil)
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeRequest(t, &input, 2, "textDocument/rename", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 2, "character": 8},
		"newName":      "Renamed",
	})
	writeRequest(t, &input, 3, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer().Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	return append([]byte(nil), output.Bytes()...)
}
