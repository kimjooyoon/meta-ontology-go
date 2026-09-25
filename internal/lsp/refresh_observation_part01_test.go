package lsp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRefreshObservationPart01TracksExactCacheReuse(t *testing.T) {
	uri := "file:///refresh-observation.gooo"
	source := "package p\nnamespace n\n"
	server := NewServer(ParserFunc(func(string, string) ParseResult {
		return ParseResult{semanticDigest: "semantic-digest"}
	}))
	params, err := json.Marshal(DidOpenTextDocumentParams{TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: source}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.didOpen(context.Background(), requestEnvelope{Params: params}); err != nil {
		t.Fatal(err)
	}
	if err := server.refresh(context.Background(), uri); err != nil {
		t.Fatal(err)
	}
	if err := server.refresh(context.Background(), uri); err != nil {
		t.Fatal(err)
	}
	observation, ok := server.RefreshObservationPart01(uri)
	if !ok {
		t.Fatal("refresh observation missing")
	}
	if !observation.ValidPart01() || observation.Decision != refreshObservationPassPart01 {
		t.Fatalf("invalid refresh observation: %#v", observation)
	}
	if observation.ParseCalls != 1 || observation.CacheHits != 1 || observation.MissingStageIndex != -1 {
		t.Fatalf("unexpected refresh counts: %#v", observation)
	}
}

func TestRefreshObservationPart01KeepsUnknownFrontier(t *testing.T) {
	uri := "file:///refresh-observation-unknown.gooo"
	source := "package p\nnamespace n\n"
	server := NewServer(ParserFunc(func(string, string) ParseResult { return ParseResult{} }))
	params, err := json.Marshal(DidOpenTextDocumentParams{TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: source}})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.didOpen(context.Background(), requestEnvelope{Params: params}); err != nil {
		t.Fatal(err)
	}
	if err := server.refresh(context.Background(), uri); err != nil {
		t.Fatal(err)
	}
	observation, ok := server.RefreshObservationPart01(uri)
	if !ok {
		t.Fatal("refresh observation missing")
	}
	if observation.Decision != refreshObservationUnknownPart01 || observation.MissingStageIndex != 1 || observation.Reason != "MISSING_SEMANTIC_DIGEST" {
		t.Fatalf("unexpected unknown frontier: %#v", observation)
	}
	if !observation.ValidPart01() {
		t.Fatalf("unknown observation should remain structurally valid: %#v", observation)
	}
}
