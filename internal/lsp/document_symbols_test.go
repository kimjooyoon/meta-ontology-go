package lsp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestDocumentSymbolRequestUsesCanonicalParseResult(t *testing.T) {
	uri := "file:///symbols.gooo"
	symbols := []Symbol{
		{ID: "urn:later", Name: "Later", Kind: SymbolClass, Detail: "entity Later", Range: testRange(2, 1, 2, 12), SelectionRange: testRange(2, 8, 2, 13)},
		{ID: "urn:first", Name: "First", Kind: SymbolClass, Detail: "entity First", Range: testRange(1, 1, 1, 12), SelectionRange: testRange(1, 8, 1, 13)},
	}
	original := append([]Symbol(nil), symbols...)
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{Symbols: append([]Symbol(nil), symbols...)}
	})
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "initialize", nil)
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "canonical"},
	})
	writeRequest(t, &input, 2, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	writeRequest(t, &input, 3, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 4 {
		t.Fatalf("output messages = %d, want 4", len(messages))
	}
	assertInitialize(t, messages[0])
	var response struct {
		ID     int              `json:"id"`
		Result []DocumentSymbol `json:"result"`
	}
	decodeJSON(t, messages[2], &response)
	if response.ID != 2 || len(response.Result) != 2 {
		t.Fatalf("document symbols response = %#v", response)
	}
	if response.Result[0].ID != "urn:first" || response.Result[1].ID != "urn:later" {
		t.Fatalf("document symbols order = %#v", response.Result)
	}
	if response.Result[0].Name != "First" || response.Result[0].Detail != "entity First" ||
		response.Result[0].Kind != SymbolClass || response.Result[0].SelectionRange.Start.Line != 1 {
		t.Fatalf("document symbol projection = %#v", response.Result[0])
	}
	if !reflect.DeepEqual(symbols, original) {
		t.Fatalf("parser symbols mutated: %#v", symbols)
	}
}

func TestCanonicalDocumentSymbolsReplayAndNoMutation(t *testing.T) {
	symbols := []Symbol{
		{ID: "z", Name: "same", Detail: "z", Kind: SymbolClass, Range: testRange(1, 1, 1, 8), SelectionRange: testRange(1, 2, 1, 6)},
		{ID: "a", Name: "same", Detail: "a", Kind: SymbolClass, Range: testRange(1, 1, 1, 8), SelectionRange: testRange(1, 2, 1, 6)},
		{ID: "early", Name: "early", Kind: SymbolFunction, Range: testRange(0, 1, 0, 8), SelectionRange: testRange(0, 2, 0, 7)},
	}
	original := append([]Symbol(nil), symbols...)
	want := canonicalDocumentSymbols(symbols)
	for replay := 0; replay < 64; replay++ {
		got := canonicalDocumentSymbols(rotateSymbols(symbols, replay))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("replay %d differs:\n got %#v\nwant %#v", replay, got, want)
		}
	}
	if !reflect.DeepEqual(symbols, original) {
		t.Fatalf("canonical projection mutated input: %#v", symbols)
	}
	if want[0].ID != "early" || want[1].ID != "a" || want[2].ID != "z" {
		t.Fatalf("canonical order = %#v", want)
	}
}

func TestDocumentSymbolInvalidAndMissingDocuments(t *testing.T) {
	uri := "file:///invalid.gooo"
	var input, output bytes.Buffer
	writeRequest(t, &input, 1, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{},
	})
	writeRequest(t, &input, 2, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": "file:///missing.gooo"},
	})
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "package p\nnamespace"},
	})
	writeRequest(t, &input, 3, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	writeRequest(t, &input, 4, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer().Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 5 {
		t.Fatalf("output messages = %d, want 5", len(messages))
	}
	if responseCode(t, messages[0]) != invalidParams {
		t.Fatalf("invalid params response code = %d", responseCode(t, messages[0]))
	}
	assertRawResult(t, messages[1], 2, "null")
	assertRawResult(t, messages[3], 3, "[]")
}

func assertRawResult(t *testing.T, payload []byte, id int, want string) {
	t.Helper()
	var response struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &response)
	if response.ID != id || string(response.Result) != want {
		t.Fatalf("response = %#v, want id=%d result=%s", response, id, want)
	}
}

func rotateSymbols(symbols []Symbol, offset int) []Symbol {
	result := make([]Symbol, len(symbols))
	for index := range symbols {
		result[index] = symbols[(index+offset)%len(symbols)]
	}
	return result
}

func testRange(startLine, startCharacter, endLine, endCharacter int) Range {
	return Range{Start: Position{Line: startLine, Character: startCharacter}, End: Position{Line: endLine, Character: endCharacter}}
}
