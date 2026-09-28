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
	if err := ValidateRepairCandidate(repair); err != nil {
		t.Fatalf("generated repair candidate did not validate: %v", err)
	}
}

func TestProposeRepairUsesSixteenDigestCharactersForCandidateID(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	candidate.Activities = []string{"Tampered"}
	repair, err := ProposeRepair(CompareReplay(baseline, candidate))
	if err != nil {
		t.Fatal(err)
	}
	const prefix = "sha256:"
	digest := repair.ComparisonDigest[len(prefix):]
	want := "gooo://repair-candidate/" + digest[:16]
	if repair.CandidateID != want {
		t.Fatalf("candidate id = %q, want %q", repair.CandidateID, want)
	}
}

func TestValidateRepairCandidateRejectsForgedIdentity(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	candidate.Activities = []string{"Tampered"}
	repair, err := ProposeRepair(CompareReplay(baseline, candidate))
	if err != nil {
		t.Fatal(err)
	}
	repair.CandidateID = "gooo://repair-candidate/forged"
	if err := ValidateRepairCandidate(repair); err == nil {
		t.Fatal("forged repair candidate was accepted")
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
