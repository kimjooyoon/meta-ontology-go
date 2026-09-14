package policycompilation

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func generatedBatchFixture(t *testing.T) (CompiledPolicy, []byte, []Case) {
	t.Helper()
	policy, err := Compile(declaredCaseSourceFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	judge := GenerateJudge(policy)
	complete := Case{
		ID: "z-pass", ValidatorExpectation: DecisionPass,
		EvidenceClass: EvidenceSyntheticFixture, Provenance: "synthetic batch fixture",
		ProducerAvailable: true, ConsumerAvailable: true,
		ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest,
		ObservedGeneratedJudgeDigest: DigestBytes(judge), ObservedIndependentDigest: policy.SemanticDigest,
	}
	unknown := Case{
		ID: "a-unknown", ValidatorExpectation: DecisionUnknown,
		EvidenceClass: EvidenceSyntheticFixture, Provenance: "synthetic missing evidence",
	}
	refuted := complete
	refuted.ID, refuted.ValidatorExpectation = "m-refuted", DecisionFailClosed
	refuted.ObservedSourceDigest = DigestBytes([]byte("synthetic different source"))
	return policy, judge, []Case{complete, unknown, refuted}
}

func TestExecuteGeneratedBatchPreservesSourceResultsAndInputOrder(t *testing.T) {
	policy, judge, inputs := generatedBatchFixture(t)
	before := append([]Case(nil), inputs...)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	results, err := ExecuteGeneratedBatch(ctx, judge, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != len(inputs) || !reflect.DeepEqual(inputs, before) {
		t.Fatalf("batch changed result count or caller observations: results=%d inputs=%d", len(results), len(inputs))
	}
	for index, input := range inputs {
		want := IndependentEvaluate(policy, input)
		if !sameResult(results[index], want) || results[index].Decision != input.ValidatorExpectation {
			t.Fatalf("case %q differs from source-derived expectation: got=%+v want=%+v", input.ID, results[index], want)
		}
	}
}

func TestExecuteGeneratedBatchRejectsEmptyInput(t *testing.T) {
	results, err := ExecuteGeneratedBatch(context.Background(), nil, nil)
	if err == nil || len(results) != 0 {
		t.Fatalf("empty batch must not claim execution: results=%v error=%v", results, err)
	}
}

func TestExecuteGeneratedBatchHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, err := ExecuteGeneratedBatch(ctx, []byte("package main\nfunc main() {}\n"), []Case{{ID: "not-started"}})
	if !errors.Is(err, context.Canceled) || len(results) != 0 {
		t.Fatalf("cancelled batch returned results or lost cancellation: results=%v error=%v", results, err)
	}
}

func TestExecuteGeneratedBatchRejectsInvalidSource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	results, err := ExecuteGeneratedBatch(ctx, []byte("package main\nfunc main( {\n"), []Case{{ID: "not-executed"}})
	if err == nil || len(results) != 0 {
		t.Fatalf("invalid generated source produced results: results=%v error=%v", results, err)
	}
}

func TestExecuteGeneratedBatchReturnsSuccessfulPrefix(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	inputs := []Case{{ID: "first"}, {ID: "malformed"}, {ID: "not-executed"}}
	results, err := ExecuteGeneratedBatch(ctx, []byte(batchProtocolFixture), inputs)
	if err == nil || len(results) != 1 || results[0].CaseID != "first" {
		t.Fatalf("batch must preserve only its successful prefix: results=%v error=%v", results, err)
	}
}

// This synthetic process-protocol fixture is not source-policy conformance.
const batchProtocolFixture = `package main

import (
	"encoding/json"
	"os"
)

func main() {
	var input map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		os.Exit(2)
	}
	if input["id"] == "malformed" {
		_, _ = os.Stdout.WriteString("{")
		return
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"case_id": input["id"]})
}
`
