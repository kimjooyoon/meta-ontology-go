package valueexecution

import "testing"

func TestPrepareAcceptedRevisionNextRunStage(t *testing.T) {
	candidate := []byte("activity ObserveRepair\n")
	comparison := AcceptedRevisionNextRunComparison{
		Schema:                       AcceptedRevisionNextRunComparisonSchema,
		State:                        ReplayClosed,
		Outcome:                      NextRunOutcomeImproved,
		Reason:                       "SOURCE_REVISION_IMPROVED_ON_NEXT_RUN",
		SourceDigest:                 "source-digest",
		CandidateSourceDigest:        digestBytes(candidate),
		RevisionCandidateID:          "candidate-1",
		Activity:                     "ObserveRepair",
		InputDigest:                  "input-digest",
		BaselineFailureCode:          "EXPECTED_FAILURE",
		BaselineExecutionDigest:      "baseline-execution",
		AcceptedExecutionDigest:      "accepted-execution",
		NextCandidateExecutionDigest: "accepted-execution",
		ExecutionAllowed:             false,
		RepositoryWrites:             0,
	}

	stage, err := PrepareAcceptedRevisionNextRunStage(comparison, candidate)
	if err != nil {
		t.Fatalf("PrepareAcceptedRevisionNextRunStage() error = %v", err)
	}
	if stage.Schema != AcceptedRevisionNextRunStageSchema || stage.Decision != AcceptedRevisionNextRunStageDecision {
		t.Fatalf("unexpected stage identity: %#v", stage)
	}
	if stage.NextOperation != AcceptedRevisionNextRunStageNextOp || stage.ExecutionAllowed || stage.RepositoryWrites != 0 {
		t.Fatalf("unexpected stage authority: %#v", stage)
	}
	if stage.ComparisonDigest == "" {
		t.Fatal("comparison digest is empty")
	}
}

func TestPrepareAcceptedRevisionNextRunStageFailsClosed(t *testing.T) {
	_, err := PrepareAcceptedRevisionNextRunStage(AcceptedRevisionNextRunComparison{
		Schema: AcceptedRevisionNextRunComparisonSchema,
		State:  ReplayUnknown,
	}, []byte("candidate"))
	if err == nil {
		t.Fatal("expected non-closed comparison to fail closed")
	}
}
