package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadMessageHandlesPartialInput(t *testing.T) {
	input := []byte("Content-Length: 11\r\n\r\n{\"ok\":true}")
	payload, err := ReadMessage(&chunkReader{data: input, size: 1})
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}
	if string(payload) != `{"ok":true}` {
		t.Fatalf("payload = %q", payload)
	}
}

func TestReadMessageRejectsMalformedHeader(t *testing.T) {
	_, err := ReadMessage(strings.NewReader("Content-Length 4\r\n\r\ntest"))
	if !errors.Is(err, ErrMalformedHeader) {
		t.Fatalf("error = %v, want malformed header", err)
	}
}

func TestReadMessageReportsPartialPayload(t *testing.T) {
	_, err := ReadMessage(strings.NewReader("Content-Length: 5\r\n\r\nabc"))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("error = %v, want unexpected EOF", err)
	}
}

func TestWriteMessageHandlesShortWrites(t *testing.T) {
	var output bytes.Buffer
	writer := &shortWriter{output: &output, size: 2}
	if err := WriteMessage(writer, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}
	payload, err := readFrame(bufio.NewReader(&output))
	if err != nil || string(payload) != `{"ok":true}` {
		t.Fatalf("framed output = %q, error = %v", payload, err)
	}
}

func TestServeReportsParseErrorAndContinues(t *testing.T) {
	var input bytes.Buffer
	writeFrameForTest(t, &input, []byte(`{"jsonrpc":"2.0",`))
	writeRequest(t, &input, 1, "shutdown", nil)
	var output bytes.Buffer
	if err := NewServer().Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 2 || responseCode(t, messages[0]) != parseError {
		t.Fatalf("messages = %s", messages)
	}
}

func TestServeExitRequiresShutdown(t *testing.T) {
	var input bytes.Buffer
	writeNotification(t, &input, "exit", nil)
	var output bytes.Buffer
	if err := NewServer().Serve(&input, &output); !errors.Is(err, ErrExitWithoutShutdown) {
		t.Fatalf("Serve() error = %v, want exit error", err)
	}
}

func writeFrameForTest(t *testing.T, output io.Writer, payload []byte) {
	t.Helper()
	if err := WriteMessage(output, payload); err != nil {
		t.Fatal(err)
	}
}

func writeRequest(t *testing.T, output io.Writer, id int, method string, params any) {
	t.Helper()
	payload := map[string]any{"jsonrpc": jsonRPCVersion, "id": id, "method": method}
	if params != nil {
		payload["params"] = params
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	writeFrameForTest(t, output, encoded)
}

func writeNotification(t *testing.T, output io.Writer, method string, params any) {
	t.Helper()
	payload := map[string]any{"jsonrpc": jsonRPCVersion, "method": method}
	if params != nil {
		payload["params"] = params
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	writeFrameForTest(t, output, encoded)
}

func readFrames(t *testing.T, data []byte) [][]byte {
	t.Helper()
	reader := bufio.NewReader(bytes.NewReader(data))
	var messages [][]byte
	for {
		message, err := readFrame(reader)
		if errors.Is(err, io.EOF) {
			return messages
		}
		if err != nil {
			t.Fatal(err)
		}
		messages = append(messages, message)
	}
}

func responseCode(t *testing.T, payload []byte) int {
	t.Helper()
	var message struct {
		Error *struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatal(err)
	}
	if message.Error == nil {
		return 0
	}
	return message.Error.Code
}

type chunkReader struct {
	data []byte
	size int
	read int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if r.read == len(r.data) {
		return 0, io.EOF
	}
	n := r.size
	if n > len(p) {
		n = len(p)
	}
	if n > len(r.data)-r.read {
		n = len(r.data) - r.read
	}
	copy(p, r.data[r.read:r.read+n])
	r.read += n
	return n, nil
}

type shortWriter struct {
	output *bytes.Buffer
	size   int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	n := w.size
	if n > len(p) {
		n = len(p)
	}
	_, _ = w.output.Write(p[:n])
	return n, nil
}
