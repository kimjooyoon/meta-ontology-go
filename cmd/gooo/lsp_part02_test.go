package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
	"testing"
)

func TestRunLSPDocumentLifecycleAndDiagnosticRequestAreProtocolOnly(t *testing.T) {
	uri := "file:///billing.gooo"
	source := "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Order) -> Order\n"
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspNotification("textDocument/didOpen", map[string]any{
			"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
		}),
		lspRequest(2, "textDocument/hover", map[string]any{
			"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 2, "character": 8},
		}),
		lspRequest(3, "shutdown", nil),
		lspNotification("exit", nil),
	)
	first, code, stderr := runLSPTranscript(t, input)
	second, secondCode, secondStderr := runLSPTranscript(t, input)
	if code != exitOK || secondCode != exitOK || stderr != "" || secondStderr != "" {
		t.Fatalf("lifecycle = %d/%d, stderr=%q/%q", code, secondCode, stderr, secondStderr)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("replayed transcript changed output:\nfirst=%s\nsecond=%s", first, second)
	}
	messages := readLSPFrames(t, first)
	if len(messages) != 4 {
		t.Fatalf("lifecycle messages = %d, want 4", len(messages))
	}
	var diagnostics struct {
		Method string                       `json:"method"`
		Params lsp.PublishDiagnosticsParams `json:"params"`
	}
	decodeLSPJSON(t, messages[1], &diagnostics)
	if diagnostics.Method != "textDocument/publishDiagnostics" || diagnostics.Params.URI != uri || len(diagnostics.Params.Diagnostics) != 0 {
		t.Fatalf("diagnostics notification = %#v", diagnostics)
	}
	var hover struct {
		Result *lsp.Hover `json:"result"`
	}
	decodeLSPJSON(t, messages[2], &hover)
	if hover.Result == nil || hover.Result.Contents.Value != "entity Order" {
		t.Fatalf("hover result = %#v", hover.Result)
	}
	assertLSPResponseID(t, messages[3], 3)
}
