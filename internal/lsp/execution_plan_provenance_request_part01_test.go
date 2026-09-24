package lsp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestExecutionPlanProvenancePreservesUnknownWithoutFullChain(t *testing.T) {
	uri := "file:///execution-plan-provenance.gooo"
	semanticDigest := cache.HashBytes([]byte("execution-plan-ir")).String()
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols:          []Symbol{{Name: "Order", ID: "order-id", SelectionRange: testRange(0, 0, 0, 5)}},
			References:       []Reference{{Name: "Order", ID: "order-id", Range: testRange(0, 6, 0, 11)}},
			semanticDigest:   semanticDigest,
			semanticChecked: true,
			semanticValid:   true,
		}
	})
	server := NewServer(parser)
	_, _, err := server.didOpen(context.Background(), requestEnvelope{
		Params: json.RawMessage(`{"textDocument":{"uri":"` + uri + `","version":1,"text":"Order Order"}}`),
	})
	if err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}
	params, err := json.Marshal(ExecutionPlanProvenanceParamsPart01{
		TextDocument:    TextDocumentIdentifier{URI: uri},
		Task:            "task-1",
		WorkspaceDigest: cache.HashBytes([]byte("workspace-1")).String(),
		Model:           "model-1",
		GatewayPolicy:   provenance.GatewayPolicy{AllowedHosts: []string{"api.example.invalid"}},
		Lifecycle:       provenance.ExecutionPlanLifecyclePlanned,
	})
	if err != nil {
		t.Fatal(err)
	}
	response, _, err := server.executionPlanProvenanceRequest(context.Background(), requestEnvelope{
		ID:     json.RawMessage("1"),
		Params: params,
	})
	if err != nil || response == nil {
		t.Fatalf("execution-plan provenance response = %#v, error = %v", response, err)
	}
	var binding provenance.ExecutionPlanProvenanceBindingPart01
	if err := json.Unmarshal(response.Result, &binding); err != nil {
		t.Fatal(err)
	}
	if binding.Schema != ExecutionPlanProvenanceSchemaPart01 ||
		binding.Status != provenance.ExecutionPlanBindingUnknown ||
		binding.BoundStages != 4 ||
		binding.TotalStages != 6 ||
		binding.MissingStageIndex != 4 ||
		binding.NextRequiredStage != "generated" ||
		binding.AdoptionAuthorized ||
		!binding.NonAuthorizing {
		t.Fatalf("unexpected execution-plan provenance binding: %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("binding.Validate() error = %v", err)
	}
	tampered := binding
	tampered.Plan.Task = "tampered"
	if err := tampered.Validate(); err == nil {
		t.Fatal("tampered execution-plan provenance binding was accepted")
	}
}
