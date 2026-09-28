package lsp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestStoryProvenanceMissingDocumentIsUnknown(t *testing.T) {
	server := &Server{documents: map[string]*document{}}
	params, err := json.Marshal(StoryProvenanceParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///missing.gooo"},
		SemanticID:   "gooo://missing",
	})
	if err != nil {
		t.Fatal(err)
	}
	response, _, err := server.storyProvenanceRequest(context.Background(), requestEnvelope{
		ID:     json.RawMessage("1"),
		Params: params,
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		t.Fatal(err)
	}
	result, ok := envelope["result"].(map[string]any)
	if !ok {
		t.Fatalf("result missing: %s", encoded)
	}
	if result["status"] != "unknown" || result["reason"] != "MISSING_DOCUMENT" {
		t.Fatalf("unexpected observation: %s", encoded)
	}
	if result["non_authorizing"] != true {
		t.Fatalf("observation must not authorize: %s", encoded)
	}
	if result["story_digest"] == "" {
		t.Fatalf("story digest missing: %s", encoded)
	}
}
