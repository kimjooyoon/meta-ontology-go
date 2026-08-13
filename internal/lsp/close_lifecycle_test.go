package lsp

import (
	"context"
	"encoding/json"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

func TestDidCloseCancelsRequestsByDocumentURI(t *testing.T) {
	methods := []string{"textDocument/hover", "textDocument/completion", "textDocument/semanticTokens/full"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) { runClosedFeature(t, method) })
	}
}

func runClosedFeature(t *testing.T, method string) {
	t.Helper()
	var calls atomic.Int32
	started, returned := make(chan struct{}), make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			return ParseResult{Symbols: []Symbol{{Name: "source", Detail: "ready"}}}, nil
		}
		close(started)
		<-ctx.Done()
		close(returned)
		return ParseResult{}, ctx.Err()
	})
	server := NewServer(parser)
	reader, writer := io.Pipe()
	output := newDiagnosticsBuffer()
	done := make(chan error, 1)
	go func() { done <- server.Serve(reader, output) }()
	uri := "file:///closed-" + method + ".gooo"
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "source"},
	})
	<-output.first
	writeRequest(t, writer, 7, method, featureRequestParams(method, uri))
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("feature parser did not start")
	}
	writeNotification(t, writer, "textDocument/didClose", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	<-output.second
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("closed feature parser did not observe cancellation")
	}
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	if responseIDPresent(t, output.Bytes(), 7) {
		t.Fatalf("closed feature emitted response: %q", output.Bytes())
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 {
		t.Fatalf("close lifecycle messages = %d, want open diagnostics, close diagnostics, shutdown", len(messages))
	}
	assertResultID(t, messages[2], 8)
	server.mu.RLock()
	_, exists := server.documents[uri]
	server.mu.RUnlock()
	if exists || calls.Load() != 2 {
		t.Fatalf("close mutated lifecycle unexpectedly: exists=%v parseCalls=%d", exists, calls.Load())
	}
}

func TestCancelRequestsForURIOnlyMatchesDocument(t *testing.T) {
	server := NewServer()
	firstCtx, firstCancel := context.WithCancel(context.Background())
	secondCtx, secondCancel := context.WithCancel(context.Background())
	server.inflight["first"] = &inFlightRequest{cancel: firstCancel, uri: "file:///first.gooo"}
	server.inflight["second"] = &inFlightRequest{cancel: secondCancel, uri: "file:///second.gooo"}
	server.cancelRequestsForURI("file:///first.gooo")
	select {
	case <-firstCtx.Done():
	default:
		t.Fatal("matching request was not canceled")
	}
	select {
	case <-secondCtx.Done():
		t.Fatal("request for another document was canceled")
	default:
	}
}

func featureRequestParams(method, uri string) map[string]any {
	params := map[string]any{"textDocument": map[string]any{"uri": uri}}
	if method != "textDocument/semanticTokens/full" {
		params["position"] = map[string]any{"character": 2}
	}
	return params
}

func responseIDPresent(t *testing.T, data []byte, id int) bool {
	t.Helper()
	for _, message := range readFrames(t, data) {
		var response struct {
			ID int `json:"id"`
		}
		if json.Unmarshal(message, &response) == nil && response.ID == id {
			return true
		}
	}
	return false
}

func TestDidChangeStaleResultUsesContentModified(t *testing.T) {
	response, _, err := featureErrorResponse(json.RawMessage("7"), ErrStaleResult, context.Background())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if responseCode(t, encoded) != contentModified {
		t.Fatalf("stale result code = %d, want %d", responseCode(t, encoded), contentModified)
	}
}
