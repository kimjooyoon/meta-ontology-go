package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"
)

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

func decodeJSON(t *testing.T, payload []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("decode %s: %v", payload, err)
	}
}

func responseCode(t *testing.T, payload []byte) int {
	t.Helper()
	var response struct {
		Error *errorObject `json:"error"`
	}
	decodeJSON(t, payload, &response)
	if response.Error == nil {
		return 0
	}
	return response.Error.Code
}

type chunkReader struct {
	data []byte
	size int
	read int
}

func (reader *chunkReader) Read(buffer []byte) (int, error) {
	if reader.read == len(reader.data) {
		return 0, io.EOF
	}
	size := reader.size
	if size > len(buffer) {
		size = len(buffer)
	}
	if size > len(reader.data)-reader.read {
		size = len(reader.data) - reader.read
	}
	copy(buffer, reader.data[reader.read:reader.read+size])
	reader.read += size
	return size, nil
}

type shortWriter struct {
	output *bytes.Buffer
	size   int
}

func (writer *shortWriter) Write(data []byte) (int, error) {
	size := writer.size
	if size > len(data) {
		size = len(data)
	}
	_, _ = writer.output.Write(data[:size])
	return size, nil
}
