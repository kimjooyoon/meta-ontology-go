package valueexecution

import (
	"math"
	"strings"
	"testing"
)

const sourceRevisionFixture = `package revision
namespace revision
entity Integer id "gooo://revision/entity/integer"
activity Observe(Integer) -> Integer computes "int.add:1"
`

func TestProposeSourceRevisionReplacesOnlyTheExactActivityProgram(t *testing.T) {
	source := []byte(sourceRevisionFixture)
	candidate, revision, err := ProposeSourceRevision("revision.gooo", source, SourceRevisionRequest{
		SourceDigest: digestBytes(source), Activity: "Observe", ExpectedProgram: "int.add:1",
		ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow,
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(candidate) == string(source) || !strings.Contains(string(candidate), `computes "int.add:0"`) {
		t.Fatalf("candidate did not replace the exact value program: %q", candidate)
	}
	if revision.ExecutionAllowed || revision.RepositoryWrites != 0 || revision.SourceDigest != digestBytes(source) || revision.CandidateSourceDigest != digestBytes(candidate) {
		t.Fatalf("revision authority = %#v", revision)
	}
	if revision.CandidateID == "" || revision.NextOperation != "EVALUATE_SOURCE_REVISION_INDEPENDENTLY" {
		t.Fatalf("revision provenance = %#v", revision)
	}
}

func TestProposeSourceRevisionFailsClosedForDigestExpectedProgramAndAmbiguity(t *testing.T) {
	cases := []SourceRevisionRequest{
		{SourceDigest: "sha256:" + strings.Repeat("0", 64), Activity: "Observe", ExpectedProgram: "int.add:1", ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow},
		{SourceDigest: digestBytes([]byte(sourceRevisionFixture)), Activity: "Observe", ExpectedProgram: "int.add:2", ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow},
	}
	for _, request := range cases {
		if _, _, err := ProposeSourceRevision("revision.gooo", []byte(sourceRevisionFixture), request); Reason(err) != ReasonSourceRevisionInvalid {
			t.Fatalf("request %#v returned %v, want %s", request, err, ReasonSourceRevisionInvalid)
		}
	}
	ambiguous := []byte(strings.Replace(sourceRevisionFixture, "activity Observe(Integer)", "activity Observe(Integer)\nactivity Observe(Integer)", 1))
	if _, _, err := ProposeSourceRevision("ambiguous.gooo", ambiguous, SourceRevisionRequest{
		SourceDigest: digestBytes(ambiguous), Activity: "Observe", ExpectedProgram: "int.add:1", ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow,
	}); Reason(err) != ReasonSourceRevisionInvalid {
		t.Fatalf("ambiguous activity returned %v, want %s", err, ReasonSourceRevisionInvalid)
	}
}

func TestEvaluateSourceRevisionRequiresKnownFailureAndAcceptsRecoveredCandidate(t *testing.T) {
	baseline := []byte(sourceRevisionFixture)
	candidate, revision, err := ProposeSourceRevision("revision.gooo", baseline, SourceRevisionRequest{
		SourceDigest: digestBytes(baseline), Activity: "Observe", ExpectedProgram: "int.add:1",
		ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow,
	})
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateSourceRevision(revision, "revision.gooo", baseline, "candidate.gooo", candidate, "Observe", math.MaxInt64)
	if evaluation.State != ReplayClosed || !evaluation.Accepted || !evaluation.CandidateExecuted || evaluation.BaselineFailure == nil || evaluation.BaselineFailure.Code != ReasonIntegerOverflow {
		t.Fatalf("accepted evaluation = %#v", evaluation)
	}
	if evaluation.RepositoryWrites != 0 || evaluation.NextOperation != "RUN_ACCEPTED_SOURCE_REVISION" || len(evaluation.BlockedBy) != 0 {
		t.Fatalf("accepted evaluation authority = %#v", evaluation)
	}

	unchanged := EvaluateSourceRevision(revision, "revision.gooo", baseline, "candidate.gooo", candidate, "Observe", 0)
	if unchanged.State != ReplayRefuted || unchanged.Reason != "SOURCE_REVISION_BASELINE_NOT_FAILED" {
		t.Fatalf("non-failing baseline = %#v", unchanged)
	}
	wrongSource := revision
	wrongSource.SourceDigest = digestBytes([]byte("other"))
	unknown := EvaluateSourceRevision(wrongSource, "revision.gooo", baseline, "candidate.gooo", candidate, "Observe", math.MaxInt64)
	if unknown.State != ReplayUnknown || unknown.Reason != "SOURCE_REVISION_EVALUATION_UNKNOWN" {
		t.Fatalf("mismatched source = %#v", unknown)
	}
}
