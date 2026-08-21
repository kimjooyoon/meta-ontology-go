package lsp

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDefinitionRequestResolvesCanonicalNamesAndIDs(t *testing.T) {
	uri := "file:///definitions.gooo"
	source := "stable_id Canonical"
	symbols := []Symbol{
		{ID: "stable_id", Name: "Canonical", SelectionRange: testRange(0, 11, 0, 20)},
	}
	original := append([]Symbol(nil), symbols...)
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{Symbols: append([]Symbol(nil), symbols...)}
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeDefinitionRequest(t, &input, 1, uri, Position{Line: 0, Character: 2})
	writeDefinitionRequest(t, &input, 2, uri, Position{Line: 0, Character: 14})
	writeDefinitionRequest(t, &input, 3, uri, Position{Line: 0, Character: 2})
	writeRequest(t, &input, 4, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 5 {
		t.Fatalf("output messages = %d, want 5", len(messages))
	}
	assertDefinitionResult(t, messages[1], 1, uri)
	assertDefinitionResult(t, messages[2], 2, uri)
	assertDefinitionResult(t, messages[3], 3, uri)
	if !reflect.DeepEqual(symbols, original) {
		t.Fatalf("definition request mutated parser symbols: %#v", symbols)
	}
}
func TestDefinitionRequestMissingAmbiguousAndCrossDocumentReturnNull(t *testing.T) {
	uri := "file:///definitions.gooo"
	source := "Dup Missing"
	symbols := []Symbol{
		{ID: "first", Name: "Dup", SelectionRange: testRange(0, 0, 0, 3)},
		{ID: "second", Name: "Dup", SelectionRange: testRange(0, 0, 0, 3)},
	}
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{Symbols: append([]Symbol(nil), symbols...)}
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeDefinitionRequest(t, &input, 1, uri, Position{Line: 0, Character: 1})
	writeDefinitionRequest(t, &input, 2, uri, Position{Line: 0, Character: 6})
	writeDefinitionRequest(t, &input, 3, "file:///other.gooo", Position{Line: 0, Character: 1})
	writeRequest(t, &input, 4, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 5 {
		t.Fatalf("output messages = %d, want 5", len(messages))
	}
	assertNullResult(t, messages[1], 1)
	assertNullResult(t, messages[2], 2)
	assertNullResult(t, messages[3], 3)
}
