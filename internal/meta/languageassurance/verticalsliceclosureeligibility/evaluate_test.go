package verticalsliceclosureeligibility

import "testing"

const testSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestEligibilitySuiteHasFixedCounts(t *testing.T) {
	suite := RunSuite(testSHA)
	if err := ValidateSuite(suite, testSHA); err != nil {
		t.Fatal(err)
	}
	if suite.CasesTotal != 8 || suite.CasesPassed != 8 || suite.CoverageBPS != 10_000 ||
		suite.EligibleExact != 1 || suite.UnknownFailClosed != 3 ||
		suite.InvariantFailClosed != 4 || suite.RepositoryWrites != 0 || suite.PromotionApplied != 0 {
		t.Fatalf("unexpected suite: %+v", suite)
	}
}

func TestExactEligibilityDoesNotActivate(t *testing.T) {
	input, _ := CaseInput("exact", testSHA)
	report := Evaluate(input)
	if err := Validate(report, input); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionEligible || report.Summary.BeforeOperating != 10 ||
		report.Summary.EligibleOperating != 11 || report.Summary.OfficialOperating != 10 ||
		report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestUnknownTopDecisionLowersResolution(t *testing.T) {
	input, _ := CaseInput("unknown-top-decision", testSHA)
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown ||
		report.EnforcementEffect != EffectBlock || report.Reason != ReasonDecisionUnknown ||
		report.Summary.UnknownPaths != 1 {
		t.Fatalf("unknown decision was not preserved: %+v", report)
	}
}

func TestEligibilityReplayIsDeterministic(t *testing.T) {
	first, second := RunSuite(testSHA), RunSuite(testSHA)
	if first.SuiteDigest != second.SuiteDigest {
		t.Fatalf("digest changed: %s != %s", first.SuiteDigest, second.SuiteDigest)
	}
}
