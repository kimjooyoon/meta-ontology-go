package lsp

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDefinitionRequestResolvesCanonicalEntityIdentityLiteral(t *testing.T) {
	uri := "file:///identity.gooo"
	source := "package p\r\nnamespace Ω\r\nentity Ω id \"urn:例/😀\"\r\n"
	identityOffset := strings.Index(source, "例")
	position, err := OffsetToPosition(source, identityOffset)
	if err != nil {
		t.Fatal(err)
	}
	nameOffset := strings.Index(source, "entity ") + len("entity ")
	nameStart, err := OffsetToPosition(source, nameOffset)
	if err != nil {
		t.Fatal(err)
	}
	nameEnd, err := OffsetToPosition(source, nameOffset+len("Ω"))
	if err != nil {
		t.Fatal(err)
	}
	identityStartOffset := strings.Index(source, "\"urn:")
	identityEndOffset := identityStartOffset + len("\"urn:例/😀\"")
	identityStart, err := OffsetToPosition(source, identityStartOffset)
	if err != nil {
		t.Fatal(err)
	}
	identityEnd, err := OffsetToPosition(source, identityEndOffset)
	if err != nil {
		t.Fatal(err)
	}

	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": source},
	})
	writeDefinitionRequest(t, &input, 1, uri, position)
	writeRequest(t, &input, 2, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	server := NewServer()
	if err := server.Serve(&input, &output); err != nil {
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
	want := Range{Start: nameStart, End: nameEnd}
	if response.ID != 1 || len(response.Result) != 1 || response.Result[0].URI != uri || response.Result[0].Range != want {
		t.Fatalf("identity definition response = %#v, want range %#v", response, want)
	}
	document := server.documents[uri]
	if document == nil || len(document.result.Symbols) != 1 {
		t.Fatalf("identity symbol = %#v", document)
	}
	symbol := document.result.Symbols[0]
	if !symbol.hasIdentity || symbol.ID != "urn:例/😀" || symbol.identityRange != (Range{Start: identityStart, End: identityEnd}) {
		t.Fatalf("identity metadata = %#v", symbol)
	}
}

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

func TestDefinitionResolutionReplayIsStable(t *testing.T) {
	symbols := []Symbol{
		{ID: "stable_id", Name: "Canonical", SelectionRange: testRange(1, 2, 1, 11)},
		{ID: "other_id", Name: "Other", SelectionRange: testRange(2, 2, 2, 7)},
	}
	want, ok := resolveDefinitionSymbol(symbols, "stable_id")
	if !ok {
		t.Fatal("stable ID did not resolve")
	}
	for replay := 0; replay < 32; replay++ {
		got, found := resolveDefinitionSymbol(rotateSymbols(symbols, replay), "stable_id")
		if !found || !reflect.DeepEqual(got, want) {
			t.Fatalf("replay %d = %#v/%v, want %#v/true", replay, got, found, want)
		}
	}
	if got, found := resolveDefinitionSymbol([]Symbol{{Name: "Canonical"}, {Name: "Canonical"}}, "Canonical"); found || got != (Symbol{}) {
		t.Fatalf("ambiguous name resolved: %#v/%v", got, found)
	}
	if got, found := resolveDefinitionSymbol([]Symbol{{ID: "stable_id"}, {ID: "stable_id"}}, "stable_id"); found || got != (Symbol{}) {
		t.Fatalf("ambiguous ID resolved: %#v/%v", got, found)
	}
}

func writeDefinitionRequest(t *testing.T, output *bytes.Buffer, id int, uri string, position Position) {
	t.Helper()
	writeRequest(t, output, id, "textDocument/definition", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": position.Line, "character": position.Character},
	})
}

func assertDefinitionResult(t *testing.T, payload []byte, id int, uri string) {
	t.Helper()
	var response struct {
		ID     int        `json:"id"`
		Result []Location `json:"result"`
	}
	decodeJSON(t, payload, &response)
	if response.ID != id || len(response.Result) != 1 || response.Result[0].URI != uri ||
		response.Result[0].Range != testRange(0, 11, 0, 20) {
		t.Fatalf("definition response = %#v", response)
	}
}

func assertNullResult(t *testing.T, payload []byte, id int) {
	t.Helper()
	var response struct {
		ID     int             `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	decodeJSON(t, payload, &response)
	if response.ID != id || string(response.Result) != "null" {
		t.Fatalf("definition null response = %#v", response)
	}
}
