package lsp

import (
	"bytes"
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
