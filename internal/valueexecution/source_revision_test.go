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

func TestSourceRevisionCarriesAnalysisProvenanceThroughEvaluation(t *testing.T) {
	source := []byte(sourceRevisionFixture)
	provenance := &AnalysisProvenance{
		SourceDigest: digestBytes(source), ProfileDigest: digestBytes([]byte("profile")),
		ToolchainDigest: digestBytes([]byte("toolchain")), ContractDigest: digestBytes([]byte("lsp-contract")),
	}
	_, revision, err := ProposeSourceRevision("revision.gooo", source, SourceRevisionRequest{
		SourceDigest: digestBytes(source), Activity: "Observe", ExpectedProgram: "int.add:1",
		ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow, AnalysisProvenance: provenance,
	})
	if err != nil {
		t.Fatal(err)
	}
	if revision.AnalysisProvenance == nil || *revision.AnalysisProvenance != *provenance {
		t.Fatalf("source revision lost analysis provenance: %#v", revision)
	}
	candidate := []byte(strings.Replace(string(source), `computes "int.add:1"`, `computes "int.add:0"`, 1))
	evaluation := EvaluateSourceRevision(revision, "revision.gooo", source, "candidate.gooo", candidate, "Observe", math.MaxInt64)
	if evaluation.AnalysisProvenance == nil || *evaluation.AnalysisProvenance != *provenance {
		t.Fatalf("evaluation lost analysis provenance: %#v", evaluation)
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

func TestEvaluateSourceRevisionSeparatesCounterexampleRecoveryFromContractAcceptance(t *testing.T) {
	baseline := []byte(sourceRevisionFixture)
	candidate, revision, err := ProposeSourceRevision("revision.gooo", baseline, SourceRevisionRequest{
		SourceDigest: digestBytes(baseline), Activity: "Observe", ExpectedProgram: "int.add:1",
		ReplacementProgram: "int.add:0", TriggerReason: ReasonIntegerOverflow,
	})
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateSourceRevision(revision, "revision.gooo", baseline, "candidate.gooo", candidate, "Observe", math.MaxInt64)
	if evaluation.State != ReplayClosed || evaluation.Accepted || !evaluation.CounterexampleRecovered || evaluation.ContractPreservation || !evaluation.CandidateExecuted || evaluation.BaselineFailure == nil || evaluation.BaselineFailure.Code != ReasonIntegerOverflow {
		t.Fatalf("accepted evaluation = %#v", evaluation)
	}
	if evaluation.RepositoryWrites != 0 || evaluation.NextOperation != "VERIFY_SOURCE_REVISION_CONTRACT" || len(evaluation.BlockedBy) == 0 {
		t.Fatalf("accepted evaluation authority = %#v", evaluation)
	}
	contract := VerifySourceRevisionContract(revision, evaluation, "revision.gooo", baseline, "candidate.gooo", candidate, "Observe", SourceRevisionContract{Scope: SourceRevisionContractScope, Inputs: []int64{0}})
	if contract.State != ReplayRefuted || contract.Accepted || contract.ContractPreservation || contract.Reason != "SOURCE_REVISION_CONTRACT_OUTPUT_CHANGED" {
		t.Fatalf("contract-changing candidate was accepted: %#v", contract)
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

func TestVerifySourceRevisionContractCanRequireIndependentExpectedOutputs(t *testing.T) {
	source := []byte(sourceRevisionFixture)
	evaluation := SourceRevisionEvaluation{State: ReplayClosed, CounterexampleRecovered: true, CandidateExecuted: true}
	contract := SourceRevisionContract{Scope: SourceRevisionContractScope, Inputs: []int64{0}, ExpectedOutputs: map[int64]int64{0: 1}}
	verified := VerifySourceRevisionContract(SourceRevision{}, evaluation, "revision.gooo", source, "candidate.gooo", source, "Observe", contract)
	if !verified.ContractPreservation || verified.ContractDigest == "" || verified.ContractExpectedOutputs[0] != 1 {
		t.Fatalf("explicit expected output was not retained as contract evidence: %#v", verified)
	}

	contract.ExpectedOutputs[0] = 2
	refuted := VerifySourceRevisionContract(SourceRevision{}, evaluation, "revision.gooo", source, "candidate.gooo", source, "Observe", contract)
	if refuted.State != ReplayRefuted || refuted.Reason != "SOURCE_REVISION_CONTRACT_BASELINE_EXPECTATION_MISMATCH" || refuted.ContractPreservation {
		t.Fatalf("baseline expectation mismatch was accepted: %#v", refuted)
	}
}
