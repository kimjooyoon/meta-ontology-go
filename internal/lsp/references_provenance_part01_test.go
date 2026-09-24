package lsp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestReferencesProvenanceBindsSemanticLocations(t *testing.T) {
	uri := "file:///references-provenance.gooo"
	semanticDigest := cache.HashBytes([]byte("references-provenance-ir")).String()
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols:        []Symbol{{Name: "Order", ID: "order-id", SelectionRange: testRange(0, 0, 0, 5)}},
			References:     []Reference{{Name: "Order", ID: "order-id", Range: testRange(0, 6, 0, 11)}},
			semanticDigest: semanticDigest, semanticChecked: true, semanticValid: true,
		}
	})
	server := NewServer(parser)
	_, _, err := server.didOpen(context.Background(), requestEnvelope{Params: json.RawMessage(`{"textDocument":{"uri":"` + uri + `","version":1,"text":"Order Order"}}`)})
	if err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}
	response, _, err := server.referencesProvenanceRequest(context.Background(), requestEnvelope{
		ID:     json.RawMessage("1"),
		Params: json.RawMessage(`{"textDocument":{"uri":"` + uri + `"},"position":{"line":0,"character":1},"context":{"includeDeclaration":true}}`),
	})
	if err != nil || response == nil {
		t.Fatalf("references provenance response = %#v, error = %v", response, err)
	}
	var value referencesProvenanceObservation
	decodeJSON(t, response.Result, &value)
	if value.Decision != referencesProvenanceClosed || value.Reason != "REFERENCE_SURFACE_BOUND" ||
		value.TargetSemanticID != "order-id" || len(value.Locations) != 2 || !value.NonAuthorizing {
		t.Fatalf("references provenance observation = %#v", value)
	}
	if value.Locations[0].Role != "declaration" || value.Locations[1].Role != "reference" {
		t.Fatalf("references provenance locations = %#v", value.Locations)
	}
	if err := validateReferencesProvenance(value); err != nil {
		t.Fatalf("validateReferencesProvenance() error = %v", err)
	}
	tampered := value
	tampered.Reason = "TAMPERED"
	if err := validateReferencesProvenance(tampered); err == nil {
		t.Fatal("validateReferencesProvenance() accepted tampered observation")
	}
}

func TestReferencesProvenancePreservesUnknownForAmbiguousTarget(t *testing.T) {
	uri := "file:///references-provenance-ambiguous.gooo"
	semanticDigest := cache.HashBytes([]byte("references-provenance-ambiguous-ir")).String()
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols:        []Symbol{{Name: "Dup"}, {Name: "Dup"}},
			semanticDigest: semanticDigest, semanticChecked: true, semanticValid: true,
		}
	})
	server := NewServer(parser)
	_, _, err := server.didOpen(context.Background(), requestEnvelope{Params: json.RawMessage(`{"textDocument":{"uri":"` + uri + `","version":1,"text":"Dup"}}`)})
	if err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}
	response, _, err := server.referencesProvenanceRequest(context.Background(), requestEnvelope{
		ID:     json.RawMessage("1"),
		Params: json.RawMessage(`{"textDocument":{"uri":"` + uri + `"},"position":{"line":0,"character":1}}`),
	})
	if err != nil || response == nil {
		t.Fatalf("references provenance response = %#v, error = %v", response, err)
	}
	var value referencesProvenanceObservation
	decodeJSON(t, response.Result, &value)
	if value.Decision != referencesProvenanceUnknown || value.Reason != "TARGET_AMBIGUOUS" || len(value.Locations) != 0 {
		t.Fatalf("ambiguous references provenance observation = %#v", value)
	}
	if err := validateReferencesProvenance(value); err != nil {
		t.Fatalf("validateReferencesProvenance() error = %v", err)
	}
}
