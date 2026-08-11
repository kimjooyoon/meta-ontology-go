package lsp

import (
	"bytes"
	"errors"
	"testing"
)

func TestReadMessageKeepsSequentialUTF8Frames(t *testing.T) {
	var input bytes.Buffer
	writeFrameForTest(t, &input, []byte(`{"text":"😀"}`))
	writeFrameForTest(t, &input, []byte(`{"text":"second"}`))

	first, err := ReadMessage(&input)
	if err != nil {
		t.Fatalf("first ReadMessage() error = %v", err)
	}
	second, err := ReadMessage(&input)
	if err != nil {
		t.Fatalf("second ReadMessage() error = %v", err)
	}
	if string(first) != `{"text":"😀"}` || string(second) != `{"text":"second"}` {
		t.Fatalf("frames = %q, %q", first, second)
	}
}

func TestInvalidNotificationDoesNotProduceResponse(t *testing.T) {
	var input bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{})
	writeNotification(t, &input, "exit", nil)
	var output bytes.Buffer
	err := NewServer().Serve(&input, &output)
	if !errors.Is(err, ErrExitWithoutShutdown) {
		t.Fatalf("Serve() error = %v, want exit error", err)
	}
	if output.Len() != 0 {
		t.Fatalf("notification produced response: %q", output.Bytes())
	}
}
