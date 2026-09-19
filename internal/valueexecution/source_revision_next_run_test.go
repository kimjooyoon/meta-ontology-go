package valueexecution

import (
	"math"
	"testing"
)

func TestCompareAcceptedRevisionNextRunClosesImprovement(t *testing.T) {
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
	accepted, err := ExecuteAcceptedSourceRevision(AcceptedSourceRevisionRequest{
		Revision: revision, Evaluation: evaluation, BaselineFilename: "baseline.gooo", BaselineSource: baseline,
		CandidateFilename: "candidate.gooo", CandidateSource: candidate, Activity: "Observe", Input: math.MaxInt64,
		ExplicitDecision: AcceptedSourceRevisionDecision,
	})
	if err != nil {
		t.Fatal(err)
	}
	comparison := CompareAcceptedRevisionNextRun(AcceptedRevisionNextRunRequest{
		Revision: revision, Evaluation: evaluation, Accepted: accepted,
		BaselineFilename: "baseline.gooo", BaselineSource: baseline,
		CandidateFilename: "candidate.gooo", CandidateSource: candidate, Activity: "Observe", Input: math.MaxInt64,
	})
	if comparison.State != ReplayClosed || comparison.Outcome != NextRunOutcomeImproved || comparison.Reason != "SOURCE_REVISION_IMPROVED_ON_NEXT_RUN" || comparison.NextOperation != "RECORD_IMPROVEMENT_EVIDENCE" || len(comparison.BlockedBy) != 0 || comparison.RepositoryWrites != 0 || comparison.AcceptedExecutionDigest != comparison.NextCandidateExecutionDigest {
		t.Fatalf("next-run comparison = %#v", comparison)
	}
}

func TestCompareAcceptedRevisionNextRunRejectsChangedCandidateIdentity(t *testing.T) {
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
	accepted, err := ExecuteAcceptedSourceRevision(AcceptedSourceRevisionRequest{
		Revision: revision, Evaluation: evaluation, BaselineFilename: "baseline.gooo", BaselineSource: baseline,
		CandidateFilename: "candidate.gooo", CandidateSource: candidate, Activity: "Observe", Input: math.MaxInt64,
		ExplicitDecision: AcceptedSourceRevisionDecision,
	})
	if err != nil {
		t.Fatal(err)
	}
	changed := append(append([]byte(nil), candidate...), '\n')
	comparison := CompareAcceptedRevisionNextRun(AcceptedRevisionNextRunRequest{
		Revision: revision, Evaluation: evaluation, Accepted: accepted,
		BaselineFilename: "baseline.gooo", BaselineSource: baseline,
		CandidateFilename: "candidate.gooo", CandidateSource: changed, Activity: "Observe", Input: math.MaxInt64,
	})
	if comparison.State != ReplayUnknown || comparison.Outcome != NextRunOutcomeUnknown || comparison.Reason != "NEXT_RUN_SCOPE_UNKNOWN" || len(comparison.BlockedBy) == 0 {
		t.Fatalf("changed candidate comparison = %#v", comparison)
	}
}
