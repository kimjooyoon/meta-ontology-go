package valueexecution

import "testing"

func replayFixture(input string) Execution {
	execution := Execution{
		Scope:       RegisteredValueOperationScope,
		PlanDigest:  digestValue("plan"),
		InputDigest: digestValue(input),
		Results:     map[string]ResultEvidence{},
		Activities:  []string{"Produce"},
	}
	execution.ExecutionDigest = executionDigest(execution)
	return execution
}

func TestCompareReplayClosesDeterministicReceipt(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	comparison := CompareReplay(baseline, candidate)
	if comparison.State != ReplayClosed || comparison.Reason != "DETERMINISTIC_REPLAY" || !comparison.SameExecution {
		t.Fatalf("comparison = %#v", comparison)
	}
}

func TestCompareReplayRejectsScopeMismatchAsUnknown(t *testing.T) {
	comparison := CompareReplay(replayFixture("one"), replayFixture("two"))
	if comparison.State != ReplayUnknown || comparison.Reason != "REPLAY_SCOPE_MISMATCH" {
		t.Fatalf("comparison = %#v", comparison)
	}
}

func TestCompareReplayRefutesChangedExecution(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	candidate.Activities = []string{"Produce", "Consume"}
	candidate.ExecutionDigest = executionDigest(candidate)
	comparison := CompareReplay(baseline, candidate)
	if comparison.State != ReplayRefuted || comparison.Reason != "REPLAY_EXECUTION_DIGEST_MISMATCH" {
		t.Fatalf("comparison = %#v", comparison)
	}
}

func TestCompareReplayFailsClosedForIncompleteReceipt(t *testing.T) {
	comparison := CompareReplay(Execution{}, replayFixture("one"))
	if comparison.State != ReplayUnknown || comparison.Reason != "REPLAY_RECEIPT_INCOMPLETE" {
		t.Fatalf("comparison = %#v", comparison)
	}
}

func TestCompareReplayRefutesReceiptContentTampering(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	candidate.Activities = []string{"Tampered"}
	comparison := CompareReplay(baseline, candidate)
	if comparison.State != ReplayRefuted || comparison.Reason != "REPLAY_RECEIPT_DIGEST_INVALID" || comparison.NextOperation != "PRESERVE_COUNTEREXAMPLE_AND_OPEN_REPAIR_CANDIDATE" {
		t.Fatalf("comparison = %#v", comparison)
	}
}
