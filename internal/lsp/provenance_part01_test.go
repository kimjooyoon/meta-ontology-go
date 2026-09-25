package lsp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDocumentProvenanceExposesExactAnalysisIdentity(t *testing.T) {
	server := NewServer()
	uri := "file:///provenance.gooo"
	source := "package provenance\nnamespace provenance\n"
	server.documents[uri] = &document{text: source}
	params, err := json.Marshal(DocumentProvenanceParams{TextDocument: TextDocumentIdentifier{URI: uri}})
	if err != nil {
		t.Fatal(err)
	}
	response, _, err := server.documentProvenanceRequest(context.Background(), requestEnvelope{ID: json.RawMessage("1"), Params: params})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || len(response.Result) == 0 {
		t.Fatalf("missing provenance response: %#v", response)
	}
	value, err := decodeDocumentProvenance(response.Result)
	if err != nil {
		t.Fatal(err)
	}
	if value.Schema != documentProvenanceSchema || value.URI != uri || value.SubjectDigest != digestText(source) || value.SourceDigest != digestText(source) || value.SemanticDigest == "" || value.ProfileDigest == "" || value.ToolchainDigest == "" || value.ContractDigest == "" || value.SymbolMapDigest == "" || value.ReferenceMapDigest == "" || len(value.Symbols) < 2 || value.ProvenanceDigest == "" || value.ProvenanceDigest != documentProvenanceDigest(value) {
		t.Fatalf("incomplete document provenance: %#v", value)
	}

	changedSource := "package provenance\nnamespace changed\n"
	changeParams, err := json.Marshal(map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": 2},
		"contentChanges": []any{map[string]any{"text": changedSource}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.didChange(context.Background(), requestEnvelope{Params: changeParams}); err != nil {
		t.Fatal(err)
	}
	response, _, err = server.documentProvenanceRequest(context.Background(), requestEnvelope{ID: json.RawMessage("2"), Params: params})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := decodeDocumentProvenance(response.Result)
	if err != nil {
		t.Fatal(err)
	}
	if changed.SourceDigest != digestText(changedSource) || changed.SourceDigest == value.SourceDigest || changed.SymbolMapDigest == value.SymbolMapDigest || changed.ProvenanceDigest == value.ProvenanceDigest {
		t.Fatalf("document provenance was stale after change: before=%#v after=%#v", value, changed)
	}
	tampered := changed
	tampered.SourceDigest = value.SourceDigest
	if validateDocumentProvenance(tampered) == nil {
		t.Fatal("tampered document provenance was accepted")
	}
}
