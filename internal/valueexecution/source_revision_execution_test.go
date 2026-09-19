package valueexecution

import (
	"math"
	"testing"
)

func TestExecuteAcceptedSourceRevisionRequiresDecisionAndBindsReexecution(t *testing.T) {
	baseline := []byte(`package revision
namespace revision
entity Integer id "revision://entity/integer"
activity Observe(Integer) -> Integer computes "int.add:1"
`)
	request := SourceRevisionRequest{
		SourceDigest: digestBytes(baseline), Activity: "Observe", ExpectedProgram: "int.add:1",
		ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow,
	}
	candidate, revision, err := ProposeSourceRevision("baseline.gooo", baseline, request)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateSourceRevision(revision, "baseline.gooo", baseline, "candidate.gooo", candidate, "Observe", math.MaxInt64)
	if evaluation.State != ReplayClosed || !evaluation.Accepted {
		t.Fatalf("evaluation = %#v", evaluation)
	}
	baseRequest := AcceptedSourceRevisionRequest{
		Revision: revision, Evaluation: evaluation, BaselineFilename: "baseline.gooo", BaselineSource: baseline,
		CandidateFilename: "candidate.gooo", CandidateSource: candidate, Activity: "Observe", Input: math.MaxInt64,
	}
	if _, err := ExecuteAcceptedSourceRevision(baseRequest); err == nil {
		t.Fatal("accepted revision executed without an explicit decision")
	}
	baseRequest.ExplicitDecision = AcceptedSourceRevisionDecision
	result, err := ExecuteAcceptedSourceRevision(baseRequest)
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != "PASS" || result.ExecutionAllowed || result.RepositoryWrites != 0 || result.NextOperation != "CAPTURE_NEXT_RUN_COMPARISON" || len(result.BlockedBy) != 0 || result.Execution.ExecutionDigest != evaluation.CandidateExecution.ExecutionDigest {
		t.Fatalf("accepted execution = %#v", result)
	}
}

func TestExecuteAcceptedSourceRevisionRejectsChangedCandidate(t *testing.T) {
	baseline := []byte(`package revision
namespace revision
entity Integer id "revision://entity/integer"
activity Observe(Integer) -> Integer computes "int.add:1"
`)
	candidate, revision, err := ProposeSourceRevision("baseline.gooo", baseline, SourceRevisionRequest{
		SourceDigest: digestBytes(baseline), Activity: "Observe", ExpectedProgram: "int.add:1",
		ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow,
	})
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateSourceRevision(revision, "baseline.gooo", baseline, "candidate.gooo", candidate, "Observe", math.MaxInt64)
	candidate = append(candidate, '\n')
	_, err = ExecuteAcceptedSourceRevision(AcceptedSourceRevisionRequest{
		Revision: revision, Evaluation: evaluation, BaselineFilename: "baseline.gooo", BaselineSource: baseline,
		CandidateFilename: "candidate.gooo", CandidateSource: candidate, Activity: "Observe", Input: math.MaxInt64,
		ExplicitDecision: AcceptedSourceRevisionDecision,
	})
	if err == nil {
		t.Fatal("changed candidate was accepted")
	}
}
