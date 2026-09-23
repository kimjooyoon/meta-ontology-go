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
	if value.Schema != documentProvenanceSchema || value.URI != uri || value.SourceDigest != digestText(source) || value.ProfileDigest == "" || value.ToolchainDigest == "" || value.ContractDigest == "" {
		t.Fatalf("incomplete document provenance: %#v", value)
	}
}
