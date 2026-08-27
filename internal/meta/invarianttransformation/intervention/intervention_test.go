package intervention

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

const testHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const testSource = `package invarianttransformation
namespace meta
entity Transformation id "gooo://invariant-transformation/value/transformation"
activity PreservedTranslation() -> Transformation computes "case=preserved-translation;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;replay=present;effect=none"
`

func TestBuildKeepsSemanticAndNonSemanticDenominatorsSeparate(t *testing.T) {
	report, err := Build([]byte(testSource), testHead)
	if err != nil {
		t.Fatal(err)
	}
	if report.Denominator.CasesTotal != 2 || report.Denominator.SemanticChange.CasesTotal != 1 || report.Denominator.NonSemantic.CasesTotal != 1 {
		t.Fatalf("denominator=%+v", report.Denominator)
	}
	if report.Denominator.SemanticChange.CasesSatisfied != 1 || report.Denominator.NonSemantic.CasesSatisfied != 1 {
		t.Fatalf("denominator satisfaction=%+v", report.Denominator)
	}
	if report.Cases[0].RawSourceDigestChanged != true || report.Cases[0].SemanticProjectionEqual || report.Cases[0].DecisionEqual || report.Cases[0].Satisfied != true {
		t.Fatalf("semantic case=%+v", report.Cases[0])
	}
	if report.Cases[1].RawSourceDigestChanged != true || !report.Cases[1].SemanticProjectionEqual || !report.Cases[1].DecisionEqual || !report.Cases[1].ResolutionEqual || !report.Cases[1].ReasonEqual || !report.Cases[1].ClaimTransitionsEqual || report.Cases[1].Satisfied != true {
		t.Fatalf("nonsemantic case=%+v", report.Cases[1])
	}
	for _, item := range report.Cases {
		if item.BaselineRepositoryWrites != 0 || item.MutatedRepositoryWrites != 0 || item.BaselineMutationAuthority || item.MutatedMutationAuthority {
			t.Fatalf("effect boundary=%+v", item)
		}
		if item.Claim.Status != model.StatusDischarged || len(item.Claim.Transitions) != 1 || item.Claim.Transitions[0].From != model.StatusOpen || item.Claim.Transitions[0].To != model.StatusDischarged || item.Claim.Transitions[0].Coordinate != item.Claim.Coordinate || item.Claim.Coordinate.Stage != InterventionStage || item.Claim.Coordinate.Reason == "" {
			t.Fatalf("claim=%+v", item.Claim)
		}
	}
}

func TestReportConsumerAndDeterministicReplayReproduceBothInterventions(t *testing.T) {
	report, err := Build([]byte(testSource), testHead)
	if err != nil {
		t.Fatal(err)
	}
	if err := DeterministicReplay(report, []byte(testSource), testHead); err != nil {
		t.Fatal(err)
	}
}
