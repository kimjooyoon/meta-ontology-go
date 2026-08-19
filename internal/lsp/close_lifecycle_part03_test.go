package lsp

import (
	"context"
	"encoding/json"
	"testing"
)

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
