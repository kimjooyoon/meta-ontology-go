package valueexecution

import "testing"

func TestProposeSourceRevisionFromHandoffBindsTriggerAndProvenance(t *testing.T) {
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
	source := []byte(sourceRevisionFixture)
	_, revision, err := ProposeSourceRevisionFromHandoff("revision.gooo", source, handoff, SourceRevisionRequest{
		SourceDigest: digestBytes(source), Activity: "Observe", ExpectedProgram: "int.add:1", ReplacementProgram: "int.add:0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if revision.TriggerReason != handoff.TriggerReason || revision.RepairCandidateID != handoff.CandidateID || revision.RepairHandoffDigest != DigestRepairHandoff(handoff) {
		t.Fatalf("revision provenance = %#v", revision)
	}
	if revision.ExecutionAllowed || revision.RepositoryWrites != 0 {
		t.Fatalf("revision granted authority = %#v", revision)
	}
}

func TestProposeSourceRevisionFromHandoffRejectsNonDeferredHandoff(t *testing.T) {
	handoff := RepairHandoff{
		Schema: RepairHandoffSchema, CandidateID: "candidate", CandidateDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ComparisonDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", TriggerState: ReplayRefuted,
		TriggerReason: "REPLAY_EXECUTION_DIGEST_MISMATCH", NextOperation: "NEXT", Status: "CLOSED",
	}
	if _, _, err := ProposeSourceRevisionFromHandoff("revision.gooo", []byte(sourceRevisionFixture), handoff, SourceRevisionRequest{}); err == nil {
		t.Fatal("non-deferred handoff was accepted")
	}
}
