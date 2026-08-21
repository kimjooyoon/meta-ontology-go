package lsp

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWriteMessageHandlesShortWrites(t *testing.T) {
	var output bytes.Buffer
	if err := WriteMessage(&shortWriter{output: &output, size: 2}, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}
	payload, err := readFrame(bufio.NewReader(&output))
	if err != nil || string(payload) != `{"ok":true}` {
		t.Fatalf("framed output = %q, error = %v", payload, err)
	}
}
