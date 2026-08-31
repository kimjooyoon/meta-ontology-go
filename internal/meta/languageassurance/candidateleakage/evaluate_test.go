package candidateleakage

import (
	"reflect"
	"testing"
)

const testSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestConformanceSuite(t *testing.T) {
	suite := RunSuite(testSHA)
	if err := ValidateSuite(suite, testSHA); err != nil {
		t.Fatal(err)
	}
	if suite.Summary.CasesTotal != 6 || suite.Summary.CasesPassed != 6 ||
		suite.Summary.ExactPass != 2 || suite.Summary.ExactFailClosed != 2 ||
		suite.Summary.InvariantFailClosed != 2 || suite.Summary.CoverageBPS != 10_000 {
		t.Fatalf("unexpected suite summary: %+v", suite.Summary)
	}
}

func TestUnknownPromotionCannotMintFixedPoint(t *testing.T) {
	input, _ := CaseInput("isolated-candidate", testSHA)
	input.Promotion.Decision = "MAYBE"
	makeOfficialPositive(&input, OfficialFixedPoint)
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionInvariant ||
		report.Reason != ReasonDecisionUnknown || report.Summary.LeakagePaths != 0 ||
		report.Summary.UnknownPaths != 1 || report.EnforcementEffect != EffectBlock {
		t.Fatalf("unknown promotion was laundered: %+v", report)
	}
}

func TestEvaluationDoesNotMutateInput(t *testing.T) {
	input, _ := CaseInput("authorized-transition", testSHA)
	before := input
	report := Evaluate(input)
	if err := Validate(report, input); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(input, before) || report.Summary.RepositoryWrites != 0 ||
		report.Summary.PromotionCreditBPS != 0 {
		t.Fatal("candidate leakage evaluation mutated its input or claimed effects")
	}
}
