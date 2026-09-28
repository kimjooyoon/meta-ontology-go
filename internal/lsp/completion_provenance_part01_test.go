package lsp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCompletionProvenanceBindsCandidatesAndRejectsTampering(t *testing.T) {
	server := NewServer()
	uri := "file:///completion-provenance.gooo"
	source := "package provenance\nnamespace provenance\n"
	server.documents[uri] = &document{text: source}
	params, err := json.Marshal(CompletionProvenanceParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 0, Character: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, _, err := server.completionProvenanceRequest(context.Background(), requestEnvelope{
		ID: json.RawMessage("1"), Params: params,
	})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || len(response.Result) == 0 {
		t.Fatalf("missing completion provenance response: %#v", response)
	}
	value, err := decodeCompletionProvenance(response.Result)
	if err != nil {
		t.Fatal(err)
	}
	if value.Schema != completionProvenanceSchema || value.URI != uri ||
		value.Decision != completionProvenanceClosed || value.Reason != "COMPLETION_SURFACE_BOUND" ||
		value.SourceDigest != digestText(source) || value.SemanticDigest == "" ||
		value.ProfileDigest == "" || value.ToolchainDigest == "" || value.ContractDigest == "" ||
		value.DocumentProvenanceDigest == "" || value.CandidateMapDigest == "" ||
		len(value.Candidates) == 0 || value.ObservationDigest == "" {
		t.Fatalf("incomplete completion provenance: %#v", value)
	}
	tampered := value
	tampered.Candidates = append([]completionProvenanceCandidate(nil), value.Candidates...)
	tampered.Candidates[0].Label = "tampered"
	if validateCompletionProvenance(tampered) == nil {
		t.Fatal("tampered completion provenance was accepted")
	}
}
