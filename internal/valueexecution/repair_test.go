package valueexecution

import "testing"

func TestProposeRepairCreatesNonExecutingCandidate(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	candidate.Activities = []string{"Tampered"}
	comparison := CompareReplay(baseline, candidate)
	repair, err := ProposeRepair(comparison)
	if err != nil {
		t.Fatal(err)
	}
	if repair.Schema != RepairCandidateSchema || repair.TriggerState != ReplayRefuted || repair.ExecutionAllowed || repair.RepositoryWrites != 0 || repair.ComparisonDigest == "" {
		t.Fatalf("repair = %#v", repair)
	}
}

func TestProposeRepairDoesNotPromoteUnknown(t *testing.T) {
	comparison := CompareReplay(Execution{}, replayFixture("one"))
	if _, err := ProposeRepair(comparison); err == nil {
		t.Fatal("unknown replay comparison became a repair candidate")
	}
}

func TestProposeRepairDoesNotPromoteClosedReplay(t *testing.T) {
	comparison := CompareReplay(replayFixture("one"), replayFixture("one"))
	if _, err := ProposeRepair(comparison); err == nil {
		t.Fatal("closed replay comparison became a repair candidate")
	}
}
