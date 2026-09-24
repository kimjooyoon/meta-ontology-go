package lsp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestSelfImprovementProvenanceChainRequestReportsNextMissingStage(t *testing.T) {
	server := NewServer()
	uri := "file:///self-improvement.gooo"
	source := "package selfimprovement\nnamespace selfimprovement\nentity Candidate id \"gooo://candidate\"\n"
	server.documents[uri] = &document{text: source}
	params, err := json.Marshal(DocumentProvenanceParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, _, err := server.selfImprovementProvenanceChainRequest(
		context.Background(),
		requestEnvelope{ID: json.RawMessage("1"), Params: params},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || len(response.Result) == 0 {
		t.Fatalf("missing self-improvement provenance response: %#v", response)
	}
	var value SelfImprovementProvenanceChainResponsePart01
	if err := json.Unmarshal(response.Result, &value); err != nil {
		t.Fatal(err)
	}
	if err := value.Validate(); err != nil {
		t.Fatal(err)
	}
	if value.Schema != SelfImprovementProvenanceChainSchemaPart01 ||
		value.URI != uri ||
		value.Status != provenance.SelfImprovementProvenanceChainUnknownPart01 ||
		value.CausalReason != "MISSING_GENERATED_DIGEST" ||
		value.NextRequiredStage != "generated" ||
		value.BoundStages != 4 ||
		value.TotalStages != 6 {
		t.Fatalf("self-improvement provenance response = %#v", value)
	}
	if value.AdoptionAuthorized || !value.NonAuthorizing {
		t.Fatalf("self-improvement provenance response crossed authorization boundary: %#v", value)
	}
}
