package intervention

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

const testHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const testSource = `package invarianttransformation
namespace meta
entity Transformation id "gooo://invariant-transformation/value/transformation"
activity PreservedTranslation() -> Transformation computes "case=preserved-translation;kind=PRESERVED;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:1;effect=none"
activity SemanticViolation() -> Transformation computes "case=semantic-violation;kind=VIOLATION;input=2;candidate=add:2;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:2;effect=none"
activity MissingRegressionWitness() -> Transformation computes "case=missing-regression-witness;kind=EVIDENCE_MISSING;input=2;candidate=add:1;expected=3;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=unavailable;effect=none"
activity ApprovedArtifact() -> Transformation computes "case=approved-artifact;kind=APPROVED_ARTIFACT;input=5;candidate=add:1;expected=6;invariant=candidate-output-equals-expected;invariant-id=candidate-output-equals-expected-v1;domain=bounded-fixture-input-domain-v1;replay=add:1;effect=approved-artifact"
`

func TestBuildKeepsInterventionDenominatorsSeparate(t *testing.T) {
	report, err := Build([]byte(testSource), testHead)
	if err != nil {
		t.Fatal(err)
	}
	if report.Denominator.CasesTotal != 3 || report.Denominator.SemanticExpectedChange.CasesTotal != 1 || report.Denominator.SemanticOperationChange.CasesTotal != 1 || report.Denominator.NonSemantic.CasesTotal != 1 {
		t.Fatalf("denominator=%+v", report.Denominator)
	}
	if report.Denominator.SemanticExpectedChange.CasesSatisfied != 1 || report.Denominator.SemanticOperationChange.CasesSatisfied != 1 || report.Denominator.NonSemantic.CasesSatisfied != 1 {
		t.Fatalf("denominator satisfaction=%+v", report.Denominator)
	}
	if report.Cases[0].RawSourceDigestChanged != true || report.Cases[0].SemanticProjectionEqual || report.Cases[0].DecisionEqual || report.Cases[0].Satisfied != true {
		t.Fatalf("semantic case=%+v", report.Cases[0])
	}
	if report.Cases[0].BaselineEvidence.ReplayCount != 2 || report.Cases[0].MutatedEvidence.ReplayCount != 2 || report.Cases[0].Claim.Coordinate.Stage != InterventionStage || report.Cases[0].Claim.Coordinate.Step != SemanticExpectedStep || report.Cases[0].Claim.Coordinate.Reason != SemanticExpectedReason {
		t.Fatalf("semantic expected evidence or claim=%+v", report.Cases[0])
	}
	if report.Cases[1].RawSourceDigestChanged != true || report.Cases[1].SemanticProjectionEqual || report.Cases[1].DecisionEqual || report.Cases[1].Satisfied != true {
		t.Fatalf("semantic operation case=%+v", report.Cases[1])
	}
	if report.Cases[1].BaselineEvidence.ReplayCount != 2 || report.Cases[1].MutatedEvidence.ReplayCount != 2 || report.Cases[1].Claim.Coordinate.Stage != InterventionStage || report.Cases[1].Claim.Coordinate.Step != SemanticOperationStep || report.Cases[1].Claim.Coordinate.Reason != SemanticOperationReason {
		t.Fatalf("semantic operation evidence or claim=%+v", report.Cases[1])
	}
	if report.Cases[2].RawSourceDigestChanged != true || !report.Cases[2].SemanticProjectionEqual || !report.Cases[2].DecisionEqual || !report.Cases[2].ResolutionEqual || !report.Cases[2].ReasonEqual || !report.Cases[2].ClaimTransitionsEqual || !report.Cases[2].EffectsEqual || !report.Cases[2].ReplayObservationEqual || report.Cases[2].Satisfied != true {
		t.Fatalf("nonsemantic case=%+v", report.Cases[2])
	}
	if report.Cases[2].BaselineEvidence.ReplayCount != 2 || report.Cases[2].MutatedEvidence.ReplayCount != 2 || report.Cases[2].Claim.Coordinate.Stage != InterventionStage || report.Cases[2].Claim.Coordinate.Step != NonSemanticStep || report.Cases[2].Claim.Coordinate.Reason != NonSemanticReason {
		t.Fatalf("nonsemantic evidence or claim=%+v", report.Cases[2])
	}
	for _, item := range report.Cases {
		if item.BaselineRepositoryWrites != -1 || item.MutatedRepositoryWrites != -1 || item.BaselineRepositoryWritesObserved || item.MutatedRepositoryWritesObserved || item.BaselineRepositoryActualOrTransientWrites != model.UnknownEffectScope || item.MutatedRepositoryActualOrTransientWrites != model.UnknownEffectScope || item.BaselineMutationAuthority || item.MutatedMutationAuthority {
			t.Fatalf("effect boundary=%+v", item)
		}
		if item.Claim.Status != model.StatusDischarged || len(item.Claim.Transitions) != 1 || item.Claim.Transitions[0].From != model.StatusOpen || item.Claim.Transitions[0].To != model.StatusDischarged || item.Claim.Transitions[0].Coordinate != item.Claim.Coordinate || item.Claim.Coordinate.Stage != InterventionStage || item.Claim.Coordinate.Reason == "" {
			t.Fatalf("claim=%+v", item.Claim)
		}
	}
	if report.EffectGateDenominator != 6 || report.EffectGateSatisfied != 6 || report.CorrectionCount != 12 || report.CorrectionDenominator != 12 {
		t.Fatalf("effect gates or corrections=%+v", report)
	}
}

func TestDeterministicReplayReproducesAllInterventions(t *testing.T) {
	report, err := Build([]byte(testSource), testHead)
	if err != nil {
		t.Fatal(err)
	}
	if err := DeterministicReplay(report, []byte(testSource), testHead); err != nil {
		t.Fatal(err)
	}
}
