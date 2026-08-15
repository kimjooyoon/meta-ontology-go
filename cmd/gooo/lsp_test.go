package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

func TestRunLSPInitializeAdvertisesExactSupportedCapabilities(t *testing.T) {
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspNotification("initialized", nil),
		lspRequest(2, "shutdown", nil),
		lspNotification("exit", nil),
	)
	first, code, stderr := runLSPTranscript(t, input)
	if code != exitOK || stderr != "" {
		t.Fatalf("initialize lifecycle = code %d, stderr=%q, output=%q", code, stderr, first)
	}
	messages := readLSPFrames(t, first)
	if len(messages) != 2 {
		t.Fatalf("lifecycle messages = %d, want 2", len(messages))
	}
	var initialize struct {
		JSONRPC string               `json:"jsonrpc"`
		ID      int                  `json:"id"`
		Result  lsp.InitializeResult `json:"result"`
	}
	decodeLSPJSON(t, messages[0], &initialize)
	if initialize.JSONRPC != "2.0" || initialize.ID != 1 || initialize.Result.ServerInfo.Name != "gooo-lsp" || initialize.Result.ServerInfo.Version != "current-ddaf" {
		t.Fatalf("initialize envelope = %#v", initialize)
	}
	want := lsp.ServerCapabilities{
		TextDocumentSync:       lsp.TextDocumentSyncOptions{OpenClose: true, Change: 2},
		HoverProvider:          true,
		CompletionProvider:     &lsp.CompletionOptions{},
		DefinitionProvider:     true,
		DocumentSymbolProvider: true,
		ReferencesProvider:     true,
		WorkspaceSymbolProvider: &lsp.WorkspaceSymbolOptions{
			Schema: lsp.WorkspaceSymbolProtocolSchema,
		},
		SemanticTokensProvider: &lsp.SemanticTokensOptions{
			Schema: lsp.SemanticTokensProtocolSchema,
			Legend: lsp.SemanticTokensLegend{
				TokenTypes:     []string{"entity", "activity", "reference", "symbol"},
				TokenModifiers: []string{},
			},
			Full: true,
		},
	}
	if !reflect.DeepEqual(initialize.Result.Capabilities, want) {
		t.Fatalf("capabilities = %#v, want %#v", initialize.Result.Capabilities, want)
	}
	assertLSPResponseID(t, messages[1], 2)
}

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

func runLSPTranscript(t *testing.T, input []byte) ([]byte, int, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := runWithInput([]string{"lsp"}, bytes.NewReader(input), &stdout, &stderr)
	return stdout.Bytes(), code, stderr.String()
}

func lspTranscript(messages ...[]byte) []byte { return bytes.Join(messages, nil) }

func lspRequest(id int, method string, params any) []byte {
	payload, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	if err != nil {
		panic(err)
	}
	return lspFrame(payload)
}

func lspNotification(method string, params any) []byte {
	payload, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
	if err != nil {
		panic(err)
	}
	return lspFrame(payload)
}

func lspFrame(payload []byte) []byte {
	return []byte(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(payload), payload))
}

func readLSPFrames(t *testing.T, data []byte) [][]byte {
	t.Helper()
	var frames [][]byte
	for len(data) > 0 {
		separator := bytes.Index(data, []byte("\r\n\r\n"))
		if separator < 0 {
			t.Fatalf("unterminated LSP header in %q", data)
		}
		header := string(data[:separator])
		if !strings.HasPrefix(header, "Content-Length: ") {
			t.Fatalf("unexpected LSP header %q", header)
		}
		length, err := strconv.Atoi(strings.TrimPrefix(header, "Content-Length: "))
		if err != nil || length < 0 || len(data) < separator+4+length {
			t.Fatalf("invalid LSP frame header %q", header)
		}
		payload := append([]byte(nil), data[separator+4:separator+4+length]...)
		if !json.Valid(payload) {
			t.Fatalf("stdout payload is not JSON-RPC: %q", payload)
		}
		frames = append(frames, payload)
		data = data[separator+4+length:]
	}
	return frames
}

func decodeLSPJSON(t *testing.T, payload []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("decode LSP payload %q: %v", payload, err)
	}
}

func assertLSPResponseID(t *testing.T, payload []byte, want int) {
	t.Helper()
	var response struct {
		ID int `json:"id"`
	}
	decodeLSPJSON(t, payload, &response)
	if response.ID != want {
		t.Fatalf("response ID = %d, want %d", response.ID, want)
	}
}

func lspResponseCode(t *testing.T, payload []byte) int {
	t.Helper()
	var response struct {
		Error struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	decodeLSPJSON(t, payload, &response)
	return response.Error.Code
}
