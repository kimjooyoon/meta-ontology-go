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

func TestReferencesRequestValidatesMissingAmbiguousAndCrossDocument(t *testing.T) {
	uri := "file:///references.gooo"
	source := "Dup Missing"
	references := []Reference{{Name: "Other", Range: testRange(0, 4, 0, 9)}}
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols: []Symbol{{Name: "Dup"}, {Name: "Dup"}}, References: references,
		}
	})
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "textDocument/references", map[string]any{"textDocument": map[string]any{}})
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeReferenceRequest(t, &input, 2, uri, Position{Line: 9, Character: 0})
	writeReferenceRequest(t, &input, 3, uri, Position{Line: 0, Character: 1})
	writeReferenceRequest(t, &input, 4, "file:///other.gooo", Position{Line: 0, Character: 1})
	writeReferenceRequest(t, &input, 5, uri, Position{Line: 0, Character: 6})
	writeRequest(t, &input, 6, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 7 {
		t.Fatalf("output messages = %d, want 7", len(messages))
	}
	if responseCode(t, messages[0]) != invalidParams || responseCode(t, messages[2]) != invalidParams {
		t.Fatalf("invalid reference responses = %d/%d", responseCode(t, messages[0]), responseCode(t, messages[2]))
	}
	assertRawReferenceResult(t, messages[3], 3, "null")
	assertRawReferenceResult(t, messages[4], 4, "null")
	assertRawReferenceResult(t, messages[5], 5, "[]")
}

func TestCanonicalReferenceLocationsReplayAndNoMutation(t *testing.T) {
	references := []Reference{
		{Name: "Other", Range: testRange(2, 0, 2, 5)},
		{Name: "Order", Range: testRange(1, 3, 1, 8)},
		{Name: "Order", Range: testRange(0, 3, 0, 8)},
	}
	original := append([]Reference(nil), references...)
	want := canonicalReferenceLocations("file:///replay.gooo", "Order", references)
	for replay := 0; replay < 32; replay++ {
		got := canonicalReferenceLocations("file:///replay.gooo", "Order", rotateReferences(references, replay))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("replay %d = %#v, want %#v", replay, got, want)
		}
	}
	if !reflect.DeepEqual(references, original) {
		t.Fatalf("canonical references mutated input: %#v", references)
	}
}

func writeReferenceRequest(t *testing.T, output *bytes.Buffer, id int, uri string, position Position) {
	t.Helper()
	writeRequest(t, output, id, "textDocument/references", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": position.Line, "character": position.Character},
		"context":      map[string]any{"includeDeclaration": true},
	})
}

func assertRawReferenceResult(t *testing.T, payload []byte, id int, want string) {
	t.Helper()
	var response struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &response)
	if response.ID != id || string(response.Result) != want {
		t.Fatalf("reference response = %#v, want id=%d result=%s", response, id, want)
	}
}

func rotateReferences(references []Reference, offset int) []Reference {
	result := make([]Reference, len(references))
	for index := range references {
		result[index] = references[(index+offset)%len(references)]
	}
	return result
}
