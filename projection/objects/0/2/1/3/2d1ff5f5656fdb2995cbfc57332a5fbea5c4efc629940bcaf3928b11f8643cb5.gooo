package lsp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestReferencesRequestUsesCurrentDocumentReferences(t *testing.T) {
	uri := "file:///references.gooo"
	source := "Order Order"
	references := []Reference{
		{Name: "Other", Range: testRange(0, 3, 0, 8)},
		{Name: "Order", Range: testRange(0, 6, 0, 11)},
		{Name: "Order", Range: testRange(0, 0, 0, 5)},
	}
	original := append([]Reference(nil), references...)
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols:    []Symbol{{Name: "Order", ID: "order-id"}},
			References: append([]Reference(nil), references...),
		}
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeReferenceRequest(t, &input, 1, uri, Position{Line: 0, Character: 1})
	writeRequest(t, &input, 2, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 {
		t.Fatalf("output messages = %d, want 3", len(messages))
	}
	var response struct {
		ID     int        `json:"id"`
		Result []Location `json:"result"`
	}
	decodeJSON(t, messages[1], &response)
	if response.ID != 1 || len(response.Result) != 2 || response.Result[0].Range != testRange(0, 0, 0, 5) ||
		response.Result[1].Range != testRange(0, 6, 0, 11) || response.Result[0].URI != uri {
		t.Fatalf("references response = %#v", response)
	}
	if !reflect.DeepEqual(references, original) {
		t.Fatalf("reference input mutated: %#v", references)
	}
}
func TestReferencesRequestRejectsMissingPosition(t *testing.T) {
	response, _, err := NewServer().referencesRequest(requestEnvelope{
		ID:     json.RawMessage("1"),
		Params: json.RawMessage(`{"textDocument":{"uri":"file:///missing-position.gooo"}}`),
	})
	if err != nil || response == nil || response.Error == nil || response.Error.Code != invalidParams {
		t.Fatalf("missing position response = %#v, error = %v", response, err)
	}
}
