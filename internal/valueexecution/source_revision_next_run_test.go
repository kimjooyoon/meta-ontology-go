package valueexecution

import (
	"math"
	"testing"
)

func TestCompareAcceptedRevisionNextRunRequiresContractPreservation(t *testing.T) {
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
	_, err = ExecuteAcceptedSourceRevision(AcceptedSourceRevisionRequest{
		Revision: revision, Evaluation: evaluation, BaselineFilename: "baseline.gooo", BaselineSource: baseline,
		CandidateFilename: "candidate.gooo", CandidateSource: candidate, Activity: "Observe", Input: math.MaxInt64,
		ExplicitDecision: AcceptedSourceRevisionDecision,
	})
	if err == nil {
		t.Fatal("counterexample recovery reached next-run execution without contract preservation")
	}
}
