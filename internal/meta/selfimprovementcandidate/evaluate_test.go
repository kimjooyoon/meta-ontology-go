package selfimprovementcandidate

import (
	"reflect"
	"testing"
)

func TestEvaluateProposesOneNonExecutingCandidate(t *testing.T) {
	head, runID := fixtureSHA("a"), int64(42)
	raw := sourceBytes(validSource(head, runID))
	first := Evaluate(validRepository(), candidateContractPath, head, runID, raw)
	second := Evaluate(validRepository(), candidateContractPath, head, runID, raw)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("candidate generation was not deterministic")
	}
	if first.Decision != DecisionProposed || first.Resolution != ResolutionExact ||
		len(first.Candidates) != 1 || first.Summary.CandidateCount != 1 ||
		first.Summary.AchievedDelta != 0 || first.Summary.TargetDelta != 1 {
		t.Fatalf("unexpected candidate report: %+v", first)
	}
	if first.Candidates[0].ExecutionAuthorized || first.Authority.RepositoryWrites != 0 {
		t.Fatalf("candidate gained authority: %+v", first)
	}
	if err := Validate(first, head, runID); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateRejectsInvalidContract(t *testing.T) {
	head, runID := fixtureSHA("b"), int64(43)
	repository := validRepository()
	file := repository[candidateContractPath]
	file.Data = []byte("package selfimprovement\nnamespace selfimprovement\n")
	report := Evaluate(repository, candidateContractPath, head, runID,
		sourceBytes(validSource(head, runID)))
	if report.Decision != DecisionFailClosed || report.Reason != ReasonContractInvalid ||
		report.Resolution != ResolutionExact || len(report.Candidates) != 0 {
		t.Fatalf("invalid contract was not closed: %+v", report)
	}
}
