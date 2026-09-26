package lsp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestObserveRefreshStageProvenancePart01UsesServerObservation(t *testing.T) {
	uri := "file:///refresh-stage-provenance.gooo"
	source := `package p
namespace n
`
	server := NewServer(ParserFunc(func(string, string) ParseResult {
		return ParseResult{semanticDigest: "semantic-digest"}
	}))
	params, err := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: source},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.didOpen(context.Background(), requestEnvelope{Params: params}); err != nil {
		t.Fatal(err)
	}

	provenance, err := server.ObserveRefreshStageProvenancePart01(context.Background(), uri)
	if err != nil {
		t.Fatal(err)
	}
	if !provenance.ValidPart01() ||
		provenance.Decision != refreshObservationPassPart01 ||
		provenance.MissingStageName != "complete" {
		t.Fatalf("invalid pass stage provenance: %#v", provenance)
	}
}

func TestObserveRefreshStageProvenancePart01KeepsUnknown(t *testing.T) {
	uri := "file:///refresh-stage-provenance-unknown.gooo"
	source := `package p
namespace n
`
	server := NewServer(ParserFunc(func(string, string) ParseResult {
		return ParseResult{}
	}))
	params, err := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Version: 1, Text: source},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.didOpen(context.Background(), requestEnvelope{Params: params}); err != nil {
		t.Fatal(err)
	}

	provenance, err := server.ObserveRefreshStageProvenancePart01(context.Background(), uri)
	if err != nil {
		t.Fatal(err)
	}
	if !provenance.ValidPart01() ||
		provenance.Decision != refreshObservationUnknownPart01 ||
		provenance.MissingStageIndex != 1 ||
		provenance.MissingStageName != "parse" {
		t.Fatalf("invalid unknown stage provenance: %#v", provenance)
	}
}
