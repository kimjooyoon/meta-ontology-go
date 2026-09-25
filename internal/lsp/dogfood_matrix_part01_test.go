package lsp

import (
	"bytes"
	"testing"
)

func TestDogfoodVersionUTF16RejectionMatrixPreservesOverlay(t *testing.T) {
	uri := "file:///dogfood-matrix.gooo"
	source := "a😀\r\nb"
	parser := ParserFunc(func(uri, source string) ParseResult { return ParseResult{} })
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
			"range": map[string]any{
				"start": map[string]any{"line": 0, "character": 1},
				"end":   map[string]any{"line": 0, "character": 2},
			},
			"text": "X",
		}},
	})
	writeNotification(t, &input, "textDocument/didChange", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 2},
		"contentChanges": []map[string]any{{
			"range": map[string]any{
				"start": map[string]any{"line": 0, "character": 1},
				"end":   map[string]any{"line": 0, "character": 3},
			},
			"text": "Z",
		}},
	})
	writeNotification(t, &input, "textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": 2},
		"contentChanges": []map[string]any{{"text": "stale-after-accept"}},
	})
	writeNotification(t, &input, "textDocument/didChange", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 3},
		"contentChanges": []map[string]any{{
			"range": map[string]any{
				"start": map[string]any{"line": 0, "character": 0},
				"end":   map[string]any{"line": 0, "character": 4},
			},
			"text": "bad",
		}},
	})
	writeRequest(t, &input, 1, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)

	server := NewServer(parser)
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	document := server.documents[uri]
	if document == nil || document.version != 2 || document.text != "aZ\r\nb" {
		t.Fatalf("dogfood matrix overlay = %#v, want version 2 and %q", document, "aZ\r\nb")
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 {
		t.Fatalf("dogfood matrix output messages = %d, want open diagnostics, accepted-change diagnostics, shutdown", len(messages))
	}
	assertDiagnostics(t, messages[0], uri, "")
	assertDiagnostics(t, messages[1], uri, "")
	assertResultID(t, messages[2], 1)
}
