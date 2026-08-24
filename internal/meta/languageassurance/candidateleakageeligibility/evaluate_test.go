package candidateleakageeligibility

import "testing"

const testSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestEligibilitySuite(t *testing.T) {
	suite := RunSuite(testSHA)
	if err := ValidateSuite(suite, testSHA); err != nil {
		t.Fatal(err)
	}
	if suite.CasesTotal != 3 || suite.CasesPassed != 3 || suite.CoverageBPS != 10_000 {
		t.Fatalf("unexpected suite: %+v", suite)
	}
}

func TestExactEligibilityHasNoSideEffect(t *testing.T) {
	input, _ := CaseInput("exact", testSHA)
	report := Evaluate(input)
	if err := Validate(report, input); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionEligible || report.Resolution != ResolutionExact ||
		report.Summary.BeforeOperating != 7 || report.Summary.AfterOperating != 8 ||
		report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		t.Fatalf("unexpected eligibility: %+v", report)
	}
}

func TestUnavailableEvidenceLowersResolution(t *testing.T) {
	input, _ := CaseInput("unavailable", testSHA)
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown ||
		report.Reason != ReasonUnavailable || report.Summary.UnknownPaths != 1 {
		t.Fatalf("unavailable evidence was laundered: %+v", report)
	}
}
