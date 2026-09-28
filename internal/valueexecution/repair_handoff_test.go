package valueexecution

import "testing"

func TestConsumeRepairCandidateCreatesDeferredHandoff(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	candidate.Activities = []string{"Tampered"}
	repair, err := ProposeRepair(CompareReplay(baseline, candidate))
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := ConsumeRepairCandidate(repair)
	if err != nil {
		t.Fatal(err)
	}
	if handoff.Schema != RepairHandoffSchema || handoff.Status != RepairHandoffDeferred || handoff.TriggerState != ReplayRefuted {
		t.Fatalf("handoff = %#v", handoff)
	}
	if handoff.CandidateDigest == "" || handoff.ComparisonDigest != repair.ComparisonDigest || handoff.NextOperation != repair.NextOperation {
		t.Fatalf("handoff provenance = %#v", handoff)
	}
	if handoff.ExecutionAllowed || handoff.RepositoryWrites != 0 {
		t.Fatalf("handoff granted authority: %#v", handoff)
	}
	if err := ValidateRepairHandoff(handoff); err != nil {
		t.Fatalf("generated handoff did not validate: %v", err)
	}
}

func TestConsumeRepairCandidateRejectsAuthorityEscalation(t *testing.T) {
	baseline := replayFixture("one")
	candidate := replayFixture("one")
	candidate.Activities = []string{"Tampered"}
	repair, err := ProposeRepair(CompareReplay(baseline, candidate))
	if err != nil {
		t.Fatal(err)
	}
	repair.ExecutionAllowed = true
	if _, err := ConsumeRepairCandidate(repair); err == nil {
		t.Fatal("authority-escalated candidate was consumed")
	}
}
