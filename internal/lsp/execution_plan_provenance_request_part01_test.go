package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestExecutionPlanProvenanceExposesTypedPlanIdentity(t *testing.T) {
	uri := "file:///typed-execution-plan-provenance.gooo"
	source := `package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"

activity ProposeCandidate(Integer) -> Integer computes "int.add:1"
activity RecordIndependentReview(Integer) -> Integer computes "int.add:1"
activity CommitCandidate(Integer) -> Integer computes "int.add:1"

bind ProposeCandidate.result -> RecordIndependentReview.input
bind RecordIndependentReview.result -> CommitCandidate.input
`
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			semanticDigest:  cache.HashBytes([]byte(source)).String(),
			semanticChecked: true,
			semanticValid:   true,
		}
	})
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(parser)
	_, _, err = server.didOpen(context.Background(), requestEnvelope{
		Params: json.RawMessage(`{"textDocument":{"uri":"` + uri + `","version":1,"text":` + string(sourceJSON) + `}}`),
	})
	if err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}
	params, err := json.Marshal(ExecutionPlanProvenanceParamsPart01{
		TextDocument:    TextDocumentIdentifier{URI: uri},
		Task:            "task-typed-plan",
		WorkspaceDigest: cache.HashBytes([]byte("workspace-typed-plan")).String(),
		Model:           "model-1",
		GatewayPolicy:   provenance.GatewayPolicy{},
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
	file, diagnostics := syntax.ParseFile(uri, source)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("typed plan parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
	if err != nil {
		t.Fatal(err)
	}
	typedPlan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		t.Fatal(err)
	}
	if binding.ProvenanceStages[3].Digest != typedPlan.Digest() {
		t.Fatalf("graph provenance digest = %q, want %q", binding.ProvenanceStages[3].Digest, typedPlan.Digest())
	}
	if binding.Plan.TypedPlanDigest != typedPlan.Digest() || binding.Plan.RuntimeBindingCount != len(typedPlan.Edges) ||
		len(binding.Plan.ActivityOrder) != len(typedPlan.Activities) || len(binding.Plan.BindingEdgeOrder) != len(typedPlan.Edges) {
		t.Fatalf("typed plan provenance = %#v, want digest %s, edges %d, activities %d", binding.Plan, typedPlan.Digest(), len(typedPlan.Edges), len(typedPlan.Activities))
	}
	for index, activity := range typedPlan.Activities {
		if binding.Plan.ActivityOrder[index] != string(activity) {
			t.Fatalf("activity order[%d] = %q, want %q", index, binding.Plan.ActivityOrder[index], activity)
		}
	}
	for index, edge := range typedPlan.Edges {
		want := executionPlanBindingEdgeKeyPart01(edge)
		if binding.Plan.BindingEdgeOrder[index] != want {
			t.Fatalf("binding edge order[%d] = %q, want %q", index, binding.Plan.BindingEdgeOrder[index], want)
		}
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("binding.Validate() error = %v", err)
	}
}

func TestExecutionPlanProvenancePreservesUnknownWithoutFullChain(t *testing.T) {
	uri := "file:///execution-plan-provenance.gooo"
	semanticDigest := cache.HashBytes([]byte("execution-plan-ir")).String()
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{
			Symbols:         []Symbol{{Name: "Order", ID: "order-id", SelectionRange: testRange(0, 0, 0, 5)}},
			References:      []Reference{{Name: "Order", ID: "order-id", Range: testRange(0, 6, 0, 11)}},
			semanticDigest:  semanticDigest,
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
		TextDocument:     TextDocumentIdentifier{URI: uri},
		Task:             "task-1",
		WorkspaceDigest:  cache.HashBytes([]byte("workspace-1")).String(),
		Model:            "model-1",
		GatewayPolicy:    provenance.GatewayPolicy{AllowedHosts: []string{"api.example.invalid"}},
		Lifecycle:        provenance.ExecutionPlanLifecyclePlanned,
		WorkloadIdentity: &provenance.WorkloadIdentityProvenanceBinding{},
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
		binding.CausalReason != "EXECUTION_PLAN_WORKLOAD_IDENTITY_INVALID" ||
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

func TestExecutionPlanProvenanceClosesOnlyCurrentGeneratedReplayEvidence(t *testing.T) {
	uri := "file:///typed-generated-replay-provenance.gooo"
	source := `package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"

activity ProposeCandidate(Integer) -> Integer computes "int.add:1"
activity RecordIndependentReview(Integer) -> Integer computes "int.add:1"

bind ProposeCandidate.result -> RecordIndependentReview.input
`
	semanticDigest := cache.HashBytes([]byte(source)).String()
	sourceDigest := "sha256:" + cache.HashBytes([]byte(source)).String()
	parser := ParserFunc(func(string, string) ParseResult {
		return ParseResult{semanticDigest: semanticDigest, semanticChecked: true, semanticValid: true}
	})
	server := NewServer(parser)
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = server.didOpen(context.Background(), requestEnvelope{
		Params: json.RawMessage(`{"textDocument":{"uri":"` + uri + `","version":1,"text":` + string(sourceJSON) + `}}`),
	})
	if err != nil {
		t.Fatalf("didOpen() error = %v", err)
	}

	file, diagnostics := syntax.ParseFile(uri, source)
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("typed plan parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
	if err != nil {
		t.Fatal(err)
	}
	typedPlan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		t.Fatal(err)
	}
	runtimePlanDigest := "sha256:" + strings.Repeat("a", 64)
	receipt := executionOriginReceiptFixturePart01(runtimePlanDigest)
	server.mu.RLock()
	stored := server.documents[uri]
	currentSemanticDigest := stored.result.semanticDigest
	server.mu.RUnlock()
	evidence := GeneratedReplayEvidencePart01{
		SourceDigest:             sourceDigest,
		SemanticDigest:           currentSemanticDigest,
		TypedPlanDigest:          typedPlan.Digest(),
		RuntimePlanDigest:        runtimePlanDigest,
		GeneratedArtifactDigest:  "sha256:" + strings.Repeat("b", 64),
		ReverseObservationDigest: "sha256:" + strings.Repeat("c", 64),
	}
	params := ExecutionPlanProvenanceParamsPart01{
		TextDocument:            TextDocumentIdentifier{URI: uri},
		Task:                    "replay-typed-plan",
		WorkspaceDigest:         "sha256:" + strings.Repeat("d", 64),
		Model:                   "model-1",
		Lifecycle:               provenance.ExecutionPlanLifecyclePlanned,
		GeneratedReplayEvidence: &evidence,
		ExecutionOriginReceipt:  &receipt,
	}
	invoke := func(id string, value ExecutionPlanProvenanceParamsPart01) provenance.ExecutionPlanProvenanceBindingPart01 {
		paramsJSON, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		response, _, requestErr := server.executionPlanProvenanceRequest(context.Background(), requestEnvelope{
			ID: json.RawMessage(id), Params: paramsJSON,
		})
		if requestErr != nil || response == nil {
			t.Fatalf("execution-plan provenance response=%#v, error=%v", response, requestErr)
		}
		var binding provenance.ExecutionPlanProvenanceBindingPart01
		if err := json.Unmarshal(response.Result, &binding); err != nil {
			t.Fatal(err)
		}
		return binding
	}

	complete := invoke("1", params)
	if complete.Status != provenance.ExecutionPlanBindingBound || complete.GeneratedReplayClosure == nil ||
		complete.GeneratedReplayClosure.Status != provenance.ExecutionEvidenceReceiptClosureCompletePart01 ||
		complete.ProvenanceStages[4].Digest != evidence.GeneratedArtifactDigest[7:] ||
		complete.ProvenanceStages[5].Digest != complete.GeneratedReplayClosure.Digest ||
		complete.AdoptionAuthorized || !complete.NonAuthorizing {
		closure := "<nil>"
		if complete.GeneratedReplayClosure != nil {
			closure = fmt.Sprintf("%+v", *complete.GeneratedReplayClosure)
		}
		t.Fatalf(
			"current generated replay was not bound as non-authorizing evidence: binding=%#v closure=%s evidence=%+v expected_typed_plan_digest=%q bound_typed_plan_digest=%q",
			complete,
			closure,
			evidence,
			typedPlan.Digest(),
			complete.Plan.TypedPlanDigest,
		)
	}

	staleEvidence := evidence
	staleEvidence.SemanticDigest = "sha256:" + strings.Repeat("e", 64)
	staleParams := params
	staleParams.GeneratedReplayEvidence = &staleEvidence
	stale := invoke("2", staleParams)
	if stale.Status != provenance.ExecutionPlanBindingUnknown || stale.GeneratedReplayClosure == nil ||
		stale.GeneratedReplayClosure.Status != provenance.ExecutionEvidenceReceiptClosureUnknownPart01 ||
		stale.GeneratedReplayClosure.MissingStageIndex != 1 ||
		stale.GeneratedReplayClosure.Reason != "GENERATED_REPLAY_SEMANTIC_DIGEST_MISMATCH" ||
		stale.MissingStageIndex != 4 || stale.NextRequiredStage != "generated" || stale.AdoptionAuthorized {
		t.Fatalf("stale generated replay evidence did not preserve UNKNOWN frontier: %#v", stale)
	}
}
