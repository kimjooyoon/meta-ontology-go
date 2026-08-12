package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemanticTokensRequestUsesCanonicalASTViews(t *testing.T) {
	uri := "file:///tokens.gooo"
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols: []Symbol{
				{Name: "Order", Kind: SymbolClass, SelectionRange: Range{Start: Position{Character: 3}, End: Position{Character: 8}}},
				{Name: "Pay", Kind: SymbolFunction, SelectionRange: Range{Start: Position{Character: 9}, End: Position{Character: 12}}},
			},
			References: []Reference{{Name: "Order", Range: Range{Start: Position{Line: 1, Character: 2}, End: Position{Line: 1, Character: 7}}}},
		}
	})
	var input, output bytes.Buffer
	writeNotification(t, &input, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "😀 Order Pay\n  Order"},
	})
	writeRequest(t, &input, 1, "textDocument/semanticTokens/full", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	writeRequest(t, &input, 2, "shutdown", nil)
	writeNotification(t, &input, "exit", nil)
	if err := NewServer(parser).Serve(&input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	messages := readFrames(t, output.Bytes())
	if len(messages) != 3 {
		t.Fatalf("output messages = %d, want diagnostics, tokens, shutdown", len(messages))
	}
	var result SemanticTokens
	decodeJSON(t, responseResult(t, messages[1]), &result)
	want := []uint32{0, 3, 5, 0, 0, 0, 6, 3, 1, 0, 1, 2, 5, 2, 0}
	if !reflect.DeepEqual(result.Data, want) {
		t.Fatalf("semantic token data = %v, want %v", result.Data, want)
	}
}

func TestSemanticTokensMalformedAndEmptyResults(t *testing.T) {
	server := NewServer()
	cases := []string{"", "null", "{}", `{"textDocument":null}`, `{"textDocument":{}}`, `{"textDocument":[]}`}
	for _, raw := range cases {
		response, _, err := server.semanticTokensRequest(context.Background(), requestEnvelope{
			ID: json.RawMessage("1"), Params: json.RawMessage(raw),
		})
		if err != nil || responseCode(t, mustMarshalResponse(t, response)) != invalidParams {
			t.Fatalf("params %q: response=%#v err=%v", raw, response, err)
		}
	}
	uri := "file:///empty.gooo"
	server.documents[uri] = &document{result: ParseResult{}}
	response, _, err := server.semanticTokensRequest(context.Background(), requestEnvelope{
		ID: json.RawMessage("2"), Params: json.RawMessage(`{"textDocument":{"uri":"file:///empty.gooo"}}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	var result SemanticTokens
	decodeJSON(t, responseResult(t, mustMarshalResponse(t, response)), &result)
	if result.Data == nil || len(result.Data) != 0 {
		t.Fatalf("empty token data = %#v, want non-nil empty", result.Data)
	}
}

func TestSemanticTokensPermutationReplayAndNoMutation(t *testing.T) {
	symbols := []Symbol{
		{Name: "later", Kind: SymbolFunction, SelectionRange: Range{Start: Position{Line: 1, Character: 4}, End: Position{Line: 1, Character: 9}}},
		{Name: "first", Kind: SymbolClass, SelectionRange: Range{Start: Position{Character: 2}, End: Position{Character: 7}}},
	}
	references := []Reference{
		{Name: "first", Range: Range{Start: Position{Line: 1, Character: 4}, End: Position{Line: 1, Character: 9}}},
		{Name: "other", Range: Range{Start: Position{Character: 10}, End: Position{Character: 15}}},
	}
	want, err := json.Marshal(semanticTokensForDocument(document{result: ParseResult{Symbols: symbols, References: references}}))
	if err != nil {
		t.Fatal(err)
	}
	originalSymbols := append([]Symbol(nil), symbols...)
	originalReferences := append([]Reference(nil), references...)
	for offset := 0; offset < len(symbols)*len(references); offset++ {
		rotatedSymbols := rotateSymbols(symbols, offset%len(symbols))
		rotatedReferences := rotateReferences(references, offset%len(references))
		got, marshalErr := json.Marshal(semanticTokensForDocument(document{result: ParseResult{
			Symbols: rotatedSymbols, References: rotatedReferences,
		}}))
		if marshalErr != nil || !bytes.Equal(got, want) {
			t.Fatalf("permutation %d: got %s want %s err=%v", offset, got, want, marshalErr)
		}
	}
	if !reflect.DeepEqual(symbols, originalSymbols) || !reflect.DeepEqual(references, originalReferences) {
		t.Fatal("semantic token projection mutated parser-owned slices")
	}
}

func TestSemanticTokensUseUTF16AndDeltaEncoding(t *testing.T) {
	start, err := OffsetToPosition("😀 Order", len("😀 "))
	if err != nil || start != (Position{Character: 3}) {
		t.Fatalf("UTF-16 start = %#v, err=%v", start, err)
	}
	result := semanticTokensForDocument(document{result: ParseResult{Symbols: []Symbol{
		{Name: "Order", Kind: SymbolClass, SelectionRange: Range{Start: start, End: Position{Character: 8}}},
		{Name: "Pay", Kind: SymbolFunction, SelectionRange: Range{Start: Position{Character: 9}, End: Position{Character: 12}}},
	}, References: []Reference{{Range: Range{Start: Position{Line: 1, Character: 2}, End: Position{Line: 1, Character: 6}}}}}})
	want := []uint32{0, 3, 5, 0, 0, 0, 6, 3, 1, 0, 1, 2, 4, 2, 0}
	if !reflect.DeepEqual(result.Data, want) {
		t.Fatalf("delta data = %v, want %v", result.Data, want)
	}
}

func TestSemanticTokensStaleOverlayReturnsContentModified(t *testing.T) {
	var calls atomic.Int32
	opened, started, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			close(opened)
			return ParseResult{Symbols: []Symbol{{Name: "source", Kind: SymbolClass, SelectionRange: Range{End: Position{Character: 6}}}}}, nil
		}
		close(started)
		select {
		case <-release:
			return ParseResult{}, nil
		case <-ctx.Done():
			return ParseResult{}, ctx.Err()
		}
	})
	reader, writer := io.Pipe()
	output := newDiagnosticsBuffer()
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, output) }()
	uri := "file:///stale-tokens.gooo"
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "source"},
	})
	<-opened
	<-output.first
	writeRequest(t, writer, 7, "textDocument/semanticTokens/full", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("semantic token parser did not start")
	}
	writeNotification(t, writer, "textDocument/didClose", map[string]any{"textDocument": map[string]any{"uri": uri}})
	<-output.second
	close(release)
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	if !semanticTokenResponseHasCode(t, output.Bytes(), 7, contentModified) {
		t.Fatalf("no ContentModified response: %q", output.Bytes())
	}
}

func TestSemanticTokensCancelSuppressesResult(t *testing.T) {
	var calls atomic.Int32
	opened, started := make(chan struct{}), make(chan struct{})
	parser := ContextParserFunc(func(ctx context.Context, uri, source string) (ParseResult, error) {
		if calls.Add(1) == 1 {
			close(opened)
			return ParseResult{}, nil
		}
		close(started)
		<-ctx.Done()
		return ParseResult{}, ctx.Err()
	})
	reader, writer := io.Pipe()
	var output bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- NewServer(parser).Serve(reader, &output) }()
	uri := "file:///cancel-tokens.gooo"
	writeNotification(t, writer, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "version": 1, "text": "source"},
	})
	<-opened
	writeRequest(t, writer, 7, "textDocument/semanticTokens/full", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	})
	<-started
	writeNotification(t, writer, "$/cancelRequest", map[string]any{"id": 7})
	writeRequest(t, writer, 8, "shutdown", nil)
	writeNotification(t, writer, "exit", nil)
	if err := <-done; err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	_ = writer.Close()
	messages := readFrames(t, output.Bytes())
	if len(messages) != 2 {
		t.Fatalf("output messages = %d, want diagnostics and shutdown", len(messages))
	}
	assertResultID(t, messages[1], 8)
}

func semanticTokenResponseHasCode(t *testing.T, data []byte, id, wantCode int) bool {
	t.Helper()
	for _, message := range readFrames(t, data) {
		var response struct {
			ID    int          `json:"id"`
			Error *errorObject `json:"error"`
		}
		decodeJSON(t, message, &response)
		if response.ID == id && response.Error != nil && response.Error.Code == wantCode {
			return true
		}
	}
	return false
}

func mustMarshalResponse(t *testing.T, response *responseEnvelope) []byte {
	t.Helper()
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
