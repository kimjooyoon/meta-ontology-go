package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestCounterexampleExecutionNativeCLIReceiptAndNextRun(t *testing.T) {
	source, input := counterexampleExecutionFixture(t)
	proposal, err := ProposePolicyRevisionFromCounterexample("policy.gooo", source,
		"metapolicycompilation", "metapolicycompilation", input.Counterexample)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	original, err := ExecuteGeneratedBatch(ctx, GenerateJudge(*proposal.OriginalPolicy), []Case{input.Counterexample.Input})
	if err != nil || len(original) != 1 {
		t.Fatalf("observe actual original result: %v", err)
	}
	input.Counterexample.Observed = original[0]
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	raw := counterexampleExecutionJSON(t, input)
	policyPath := writeRevisionOperationInput(t, work, "policy.gooo", source)
	inputPath := writeRevisionOperationInput(t, work, "input.json", raw)
	contract := PolicyRevisionOperationContract()
	operationPath := writeRevisionOperationInput(t, work, "operation.gooo", contract)
	output := runRevisionOperationNativeCLI(t, root, "./examples/meta-policy-compilation/counterexample-execution",
		"-policy", policyPath, "-input", inputPath, "-operation", operationPath)
	var report PolicyCounterexampleExecution
	if err := json.Unmarshal(output, &report); err != nil {
		t.Fatal(err)
	}
	assertCounterexampleExecutionWitness(t, report, input, raw)
	requestPath := writeRevisionOperationInput(t, work, "request.json", []byte(report.RevisionRequestJSON))
	receiptPath := writeRevisionOperationInput(t, work, "receipt.json", counterexampleExecutionJSON(t, report.Operation.Observation))
	independent := runRevisionOperationNativeCLI(t, root, "./cmd/meta-policy-compilation-consumer",
		"-policy", policyPath, "-revision-request", requestPath, "-observe-revision-receipt", receiptPath)
	assertRevisionOperationIndependentReceipt(t, independent, len(input.Cases))
	observeCounterexampleFreshNextRun(t, ctx, report, input.Cases[0].Candidate)
	for path, before := range map[string][]byte{policyPath: source, inputPath: raw, operationPath: contract} {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("caller input changed: %s error=%v", path, err)
		}
	}
}

func assertCounterexampleExecutionWitness(t *testing.T, report PolicyCounterexampleExecution, input PolicyCounterexampleExecutionInput, raw []byte) {
	t.Helper()
	if report.Schema != PolicyCounterexampleExecutionSchema || report.InputArtifactDigest != DigestBytes(raw) ||
		report.Proposal.State != "PROPOSED" || !reflect.DeepEqual(report.Proposal.Counterexample, input.Counterexample) ||
		report.Operation == nil || report.Operation.NativeWorkerInvocations != 1 || report.Operation.Observation == nil {
		t.Fatalf("missing exact counterexample execution: %+v", report)
	}
	operation, observed := report.Operation, report.Operation.Observation
	if !operation.Binding.UsedPolicySource || !operation.Binding.UsedRevisionRequest || !operation.Binding.GeneratedObservation ||
		operation.RequestArtifactDigest != DigestBytes([]byte(report.RevisionRequestJSON)) ||
		observed.Request.Condition != input.Counterexample.Observed.MatchedCondition ||
		observed.Request.FromDecision != input.Counterexample.Observed.Decision ||
		observed.Request.ToDecision != input.Counterexample.Input.ValidatorExpectation ||
		!reflect.DeepEqual(observed.Request.Cases, input.Cases) || observed.CandidateSource != report.Proposal.CandidateSource {
		t.Fatal("Gooo binding or derived request differs from caller input")
	}
	want := PolicyRevisionObservationCounts{
		RequestedCasePairs: 3, ObservedCasePairs: 3, SameInputObservedPairs: 2,
		RequestedTransitionsObserved: 1, SourceComparisons: 12, ReplayComparisons: 6,
	}
	if observed.Counts != want || observed.ExecutionConformance != "PASS" || observed.ExecutionStatus != "COMPLETED" ||
		observed.Admission.State != "UNKNOWN" || observed.Improvement != "UNKNOWN" ||
		report.Proposal.Admission.State != "UNKNOWN" || report.Proposal.CandidateExecution != "NOT_OBSERVED" ||
		operation.MutationAuthority != 0 || operation.PromotionAuthority != 0 {
		t.Fatalf("stages, exact counts or authority changed: %+v", report)
	}
}

func observeCounterexampleFreshNextRun(t *testing.T, ctx context.Context, report PolicyCounterexampleExecution, input Case) {
	t.Helper()
	observed := report.Operation.Observation
	judge := []byte(observed.Candidate.GeneratedJudgeSource)
	if DigestBytes(judge) != observed.Candidate.GeneratedJudgeDigest ||
		DigestBytes([]byte(observed.CandidateSource)) != observed.CandidatePolicy.SourceDigest {
		t.Fatal("next run lost candidate source or generated artifact identity")
	}
	results, err := ExecuteGeneratedBatch(ctx, judge, []Case{input})
	if err != nil || len(results) != 1 || results[0].Decision != input.ValidatorExpectation ||
		results[0].MatchedCondition != observed.Request.Condition ||
		results[0].PolicyDigest != observed.CandidatePolicy.SourceDigest ||
		results[0].SemanticDigest != observed.CandidatePolicy.SemanticDigest {
		t.Fatalf("fresh run did not consume the candidate artifact: %+v error=%v", results, err)
	}
	t.Logf("COUNTEREXAMPLE_EXECUTION_WITNESS=%s", counterexampleExecutionJSON(t, map[string]any{
		"evidence_class": EvidenceSyntheticFixture, "activity": report.Operation.Binding.ActivityID,
		"input_artifact_digest": report.InputArtifactDigest, "counts": observed.Counts,
		"native_worker_invocations": report.Operation.NativeWorkerInvocations,
		"next_run_generated_digest": DigestBytes(judge), "next_run_results": results,
		"admission": observed.Admission, "improvement": observed.Improvement,
	}))
}
