package lsp

import (
	"encoding/json"
	"strings"
	"testing"
)

func responseResult(t *testing.T, payload []byte) json.RawMessage {
	t.Helper()
	var response struct {
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &response)
	return response.Result
}
func assertDiagnostics(t *testing.T, payload []byte, uri, code string) {
	t.Helper()
	var message struct {
		Method string                   `json:"method"`
		Params PublishDiagnosticsParams `json:"params"`
	}
	decodeJSON(t, payload, &message)
	if message.Method != "textDocument/publishDiagnostics" || message.Params.URI != uri {
		t.Fatalf("diagnostic notification = %#v", message)
	}
	if code == "" && len(message.Params.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics = %#v", message.Params.Diagnostics)
	}
	if code != "" && (len(message.Params.Diagnostics) != 1 || message.Params.Diagnostics[0].Code != code) {
		t.Fatalf("diagnostics = %#v", message.Params.Diagnostics)
	}
}
func assertHover(t *testing.T, payload []byte, want string) {
	t.Helper()
	var message struct {
		Result *Hover `json:"result"`
	}
	decodeJSON(t, payload, &message)
	if message.Result == nil || !strings.Contains(message.Result.Contents.Value, want) {
		t.Fatalf("hover = %#v", message.Result)
	}
}
func assertCompletion(t *testing.T, payload []byte, want string) {
	t.Helper()
	var message struct {
		Result CompletionList `json:"result"`
	}
	decodeJSON(t, payload, &message)
	for _, item := range message.Result.Items {
		if item.Label == want {
			return
		}
	}
	t.Fatalf("completion items = %#v", message.Result.Items)
}
func assertDefinition(t *testing.T, payload []byte, uri string) {
	t.Helper()
	var message struct {
		Result []Location `json:"result"`
	}
	decodeJSON(t, payload, &message)
	if len(message.Result) != 1 || message.Result[0].URI != uri {
		t.Fatalf("definition = %#v", message.Result)
	}
}
