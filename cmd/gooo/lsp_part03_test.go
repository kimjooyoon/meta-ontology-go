package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
	"testing"
)

func TestRunLSPStaleVersionPreservesPriorOverlay(t *testing.T) {
	uri := "file:///version.gooo"
	source := "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\n"
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspNotification("textDocument/didOpen", map[string]any{
			"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
		}),
		lspRequest(4, "textDocument/didChange", map[string]any{
			"textDocument":   map[string]any{"uri": uri, "version": 1},
			"contentChanges": []map[string]any{{"text": "stale"}},
		}),
		lspRequest(5, "textDocument/hover", map[string]any{
			"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 2, "character": 8},
		}),
		lspRequest(6, "shutdown", nil),
		lspNotification("exit", nil),
	)
	messages, code, stderr := runLSPTranscript(t, input)
	if code != exitOK || stderr != "" {
		t.Fatalf("stale-version lifecycle = code %d, stderr=%q", code, stderr)
	}
	frames := readLSPFrames(t, messages)
	if len(frames) != 5 {
		t.Fatalf("stale-version messages = %d, want 5", len(frames))
	}
	if got := lspResponseCode(t, frames[2]); got != -32602 {
		t.Fatalf("stale-version code = %d, want -32602", got)
	}
	var hover struct {
		Result *lsp.Hover `json:"result"`
	}
	decodeLSPJSON(t, frames[3], &hover)
	if hover.Result == nil || hover.Result.Contents.Value != "entity Order" {
		t.Fatalf("stale overlay hover = %#v", hover.Result)
	}
	assertLSPResponseID(t, frames[4], 6)
}
func TestRunLSPUnknownAndMalformedRequestsFailThroughProtocol(t *testing.T) {
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspFrame([]byte("{")),
		lspRequest(2, "textDocument/rename", nil),
		lspRequest(3, "workspace/unknown", nil),
		lspRequest(4, "shutdown", nil),
		lspNotification("exit", nil),
	)
	output, code, stderr := runLSPTranscript(t, input)
	if code != exitOK || stderr != "" {
		t.Fatalf("unsupported lifecycle = code %d, stderr=%q, output=%q", code, stderr, output)
	}
	frames := readLSPFrames(t, output)
	if len(frames) != 5 {
		t.Fatalf("unsupported messages = %d, want 5", len(frames))
	}
	if got := lspResponseCode(t, frames[1]); got != -32700 {
		t.Fatalf("malformed JSON code = %d, want -32700", got)
	}
	if got := lspResponseCode(t, frames[2]); got != -32601 {
		t.Fatalf("deferred method code = %d, want -32601", got)
	}
	if got := lspResponseCode(t, frames[3]); got != -32601 {
		t.Fatalf("unknown method code = %d, want -32601", got)
	}
	assertLSPResponseID(t, frames[4], 4)
}
