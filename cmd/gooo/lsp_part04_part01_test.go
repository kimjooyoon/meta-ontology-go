package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

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
