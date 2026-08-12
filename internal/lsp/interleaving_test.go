package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type diagnosticsBuffer struct {
	mu     sync.Mutex
	data   bytes.Buffer
	count  int
	first  chan struct{}
	second chan struct{}
}

func newDiagnosticsBuffer() *diagnosticsBuffer {
	return &diagnosticsBuffer{first: make(chan struct{}), second: make(chan struct{})}
}

func (buffer *diagnosticsBuffer) Write(data []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	_, _ = buffer.data.Write(data)
	if bytes.Contains(data, []byte(`"method":"textDocument/publishDiagnostics"`)) {
		buffer.count++
		if buffer.count == 1 {
			close(buffer.first)
		}
		if buffer.count == 2 {
			close(buffer.second)
		}
	}
	return len(data), nil
}

func (buffer *diagnosticsBuffer) Bytes() []byte {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return append([]byte(nil), buffer.data.Bytes()...)
}

type trackedPipeReader struct {
	reader  *io.PipeReader
	started chan struct{}
	once    sync.Once
}

func (reader *trackedPipeReader) Read(data []byte) (int, error) {
	reader.once.Do(func() { close(reader.started) })
	return reader.reader.Read(data)
}

func (reader *trackedPipeReader) Close() error {
	return reader.reader.Close()
}

func TestServeContextClosesBlockedInputRead(t *testing.T) {
	pipeReader, pipeWriter := io.Pipe()
	input := &trackedPipeReader{reader: pipeReader, started: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- NewServer().ServeContext(ctx, input, &bytes.Buffer{}) }()
	<-input.started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ServeContext() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ServeContext() did not unblock its input read")
	}
	if _, err := pipeWriter.Write([]byte("unexpected input")); err == nil {
		t.Fatal("pipe writer remained open after canceled input read")
	}
	_ = pipeWriter.Close()
}

func TestExitCancelsPendingRequest(t *testing.T) {
	var calls atomic.Int32
	opened := make(chan struct{})
	started := make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			close(opened)
			return ParseResult{Symbols: []Symbol{{Name: "source", Detail: "ready"}}}, nil
		}
		close(started)
		<-ctx.Done()
		return ParseResult{}, ctx.Err()
	})
	reader, writer := io.Pipe()
	output := newDiagnosticsBuffer()
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, output) }()
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": "file:///exit.gooo", "version": 1, "text": "source"},
	})
	<-opened
	writeRequest(t, writer, 7, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": "file:///exit.gooo"}, "position": map[string]any{"character": 2},
	})
	<-started
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	messages := readFrames(t, output.Bytes())
	if len(messages) != 2 {
		t.Fatalf("output messages = %d, want diagnostics and shutdown only", len(messages))
	}
	assertResultID(t, messages[1], 8)
}

func TestStaleFeatureReturnsContentModified(t *testing.T) {
	var calls atomic.Int32
	opened := make(chan struct{})
	release := make(chan struct{})
	returned := make(chan struct{})
	started := make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			close(opened)
			return ParseResult{Symbols: []Symbol{{Name: "source", Detail: "ready"}}}, nil
		}
		close(started)
		select {
		case <-release:
			close(returned)
			return ParseResult{Symbols: []Symbol{{Name: "source", Detail: "new"}}}, nil
		case <-ctx.Done():
			return ParseResult{}, ctx.Err()
		}
	})
	reader, writer := io.Pipe()
	output := newDiagnosticsBuffer()
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, output) }()
	uri := "file:///stale.gooo"
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "source"},
	})
	<-opened
	<-output.first
	writeRequest(t, writer, 7, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"character": 2},
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("stale feature parser did not start")
	}
	writeNotification(t, writer, "textDocument/didClose", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	<-output.second
	close(release)
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("stale feature parser did not return")
	}
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	messages := readFrames(t, output.Bytes())
	for _, message := range messages {
		var response struct {
			ID    int          `json:"id"`
			Error *errorObject `json:"error"`
		}
		if json.Unmarshal(message, &response) != nil || response.ID != 7 || response.Error == nil {
			continue
		}
		if response.Error.Code == contentModified && response.Error.Message == "content modified during request" {
			return
		}
	}
	t.Fatalf("messages contained no ContentModified response: %q", output.Bytes())
}
