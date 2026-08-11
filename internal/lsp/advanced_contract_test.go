package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestSyntaxParserReportsSpannedDiagnostic(t *testing.T) {
	source := "package billing\r\nnamespace billing\r\nentity Order id \"unterminated"
	result := (SyntaxParser{}).Parse("file:///billing.gooo", source)
	if len(result.Diagnostics) == 0 {
		t.Fatal("Parse() returned no diagnostics")
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Code != "lex.unterminated-string" || diagnostic.Source != "gooo" {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	wantStart := Position{Line: 2, Character: 16}
	if diagnostic.Range.Start != wantStart || diagnostic.Range.End.Character <= wantStart.Character {
		t.Fatalf("diagnostic range = %#v, want start %#v and a spanning end", diagnostic.Range, wantStart)
	}
}

func TestUTF16PositionsClampCRLF(t *testing.T) {
	source := "a😀b\r\nnext"
	if got := offsetPosition(source, len("a😀")); got != (Position{Line: 0, Character: 3}) {
		t.Fatalf("offsetPosition() = %#v, want line 0 UTF-16 character 3", got)
	}
	if got := positionOffset(source, Position{Line: 0, Character: 99}); got != len("a😀b") {
		t.Fatalf("positionOffset() = %d, want CR boundary %d", got, len("a😀b"))
	}
	if got := positionOffset(source, Position{Line: 1}); got != len("a😀b\r\n") {
		t.Fatalf("second-line offset = %d, want %d", got, len("a😀b\r\n"))
	}
}

func TestDocumentOverlaysRemainURILocal(t *testing.T) {
	parser := ParserFunc(func(uri, source string) ParseResult {
		return ParseResult{Diagnostics: []Diagnostic{{Message: uri + ":" + source}}}
	})
	server := NewServer(parser)
	uriA, uriB := "file:///a.gooo", "file:///b.gooo"
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uriA, "version": 1, "text": "A"},
	})
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uriB, "version": 1, "text": "B"},
	})
	writeNotification(t, &input, "textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uriA, "version": 2},
		"contentChanges": []map[string]any{{"text": "A2"}},
	})
	writeNotification(t, &input, "textDocument/didClose", map[string]any{
		"textDocument": map[string]any{"uri": uriA},
	})
	writeRequest(t, &input, 1, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := server.Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	if _, ok := server.documents[uriA]; ok {
		t.Fatal("closed overlay remains in the document map")
	}
	document, ok := server.documents[uriB]
	if !ok || document.text != "B" || document.version != 1 {
		t.Fatalf("URI B overlay = %#v, want version 1 text B", document)
	}
	if messages := readFrames(t, output.Bytes()); len(messages) != 5 {
		t.Fatalf("output messages = %d, want 5", len(messages))
	}
}

func TestSymbolAliasesResolveToCanonicalFeatures(t *testing.T) {
	uri := "file:///aliases.gooo"
	symbol := Symbol{
		Name: "Order", Aliases: []string{"Purchase"}, ID: "billing://entity/order",
		Kind: symbolClass, Detail: "billing://entity/order",
		SelectionRange: Range{Start: Position{}, End: Position{Character: 5}},
	}
	server := &Server{documents: map[string]*document{
		uri: {text: "Purchase", result: ParseResult{Symbols: []Symbol{symbol}}},
	}}
	params := TextDocumentPositionParams{TextDocument: TextDocumentIdentifier{URI: uri}, Position: Position{Character: 3}}
	hover, ok := server.hover(params)
	if !ok || hover.Contents.Value != "billing://entity/order" {
		t.Fatalf("hover = %#v, found = %v", hover, ok)
	}
	locations := server.definition(params)
	if len(locations) != 1 || locations[0].Range.End.Character != 5 {
		t.Fatalf("definitions = %#v", locations)
	}
	items := server.completion(uri).Items
	for _, item := range items {
		if item.Label == "Purchase" && item.Detail == "alias of Order" {
			return
		}
	}
	t.Fatalf("alias completion missing: %#v", items)
}

func TestRequestResponseRetainsStringID(t *testing.T) {
	var input, output bytes.Buffer
	writeFrameForTest(t, &input, []byte(`{"jsonrpc":"2.0","id":"request-1","method":"shutdown"}`))
	writeNotification(t, &input, "exit", nil)
	if err := NewServer().Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	var response struct {
		ID     json.RawMessage `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	frames := readFrames(t, output.Bytes())
	if len(frames) != 1 {
		t.Fatalf("response frames = %d, want 1", len(frames))
	}
	decode(t, frames[0], &response)
	if string(response.ID) != `"request-1"` || string(response.Result) != "null" {
		t.Fatalf("response = %#v", response)
	}
}

func TestContextParserStopsOnCancellation(t *testing.T) {
	started := make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		close(started)
		<-ctx.Done()
		return ParseResult{}, ctx.Err()
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": "file:///cancel.gooo", "version": 1, "text": "source"},
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).ServeContext(ctx, &input, &output) }()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("ServeContext() error = %v, want context.Canceled", err)
	}
	if output.Len() != 0 {
		t.Fatalf("canceled parse produced output: %q", output.Bytes())
	}
}
