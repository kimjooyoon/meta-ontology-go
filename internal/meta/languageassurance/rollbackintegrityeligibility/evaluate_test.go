package rollbackintegrityeligibility

import "testing"

const testSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestEligibilitySuite(t *testing.T) {
	suite := RunSuite(testSHA)
	if err := ValidateSuite(suite, testSHA); err != nil {
		t.Fatal(err)
	}
	if suite.CasesTotal != 4 || suite.CasesPassed != 4 || suite.CoverageBPS != 10_000 {
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
		report.Summary.BeforeOperating != 9 || report.Summary.AfterOperating != 10 ||
		report.Summary.CapsulesExact != 3 || report.Summary.ShadowReplaysExact != 2 ||
		report.Summary.MetaOperationsObserved != 1 || len(report.Indicators) != 6 ||
		report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		t.Fatalf("unexpected eligibility: %+v", report)
	}
}

func TestUnavailableEvidenceLowersResolution(t *testing.T) {
	input, _ := CaseInput("unavailable", testSHA)
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown ||
		report.Reason != ReasonUnavailable || report.EnforcementEffect != EffectBlock ||
		report.Summary.UnknownPaths != 1 || report.Summary.AfterOperating != 9 {
		t.Fatalf("unavailable evidence was laundered: %+v", report)
	}
}

func TestDigestMismatchFailsInvariantOnly(t *testing.T) {
	input, _ := CaseInput("digest-mismatch", testSHA)
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionInvariant ||
		report.Reason != ReasonDigestMismatch || report.EnforcementEffect != EffectBlock ||
		report.Summary.AfterOperating != 9 || report.PromotionApplied != 0 {
		t.Fatalf("digest mismatch was accepted: %+v", report)
	}
}

func TestUnknownSubjectFailsClosed(t *testing.T) {
	input, _ := CaseInput("subject-unknown", testSHA)
	report := Evaluate(input)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown ||
		report.Reason != ReasonSubjectUnknown || report.Summary.UnknownPaths != 1 ||
		report.Summary.AfterOperating != 9 || report.PromotionApplied != 0 {
		t.Fatalf("unknown subject was promoted: %+v", report)
	}
}
