package changedsurfacereceipt

import (
	"reflect"
	"testing"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestSuiteIsExactAndReplayable(t *testing.T) {
	first, second := EvaluateSuite(candidateSHA), EvaluateSuite(candidateSHA)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("suite replay differs")
	}
	if first.Decision != DecisionFixedPoint || first.Resolution != ResolutionExact || first.CasesTotal != 6 || first.CasesPassed != 6 || first.CoverageBPS != 10000 {
		t.Fatalf("suite=%+v", first)
	}
}

func TestExactReceiptTotalityHasNoEffects(t *testing.T) {
	report := Evaluate(CaseInput("exact", candidateSHA))
	if report.Decision != DecisionFixedPoint || report.Resolution != ResolutionExact || report.Summary.TotalityBPS != 10000 || report.RepositoryWrites != 0 {
		t.Fatalf("report=%+v", report)
	}
	counts := map[string]int{}
	for _, indicator := range report.Indicators {
		if !indicator.Satisfied {
			t.Fatalf("indicator=%+v", indicator)
		}
		counts[indicator.Class]++
		counts[indicator.ProofChoice]++
	}
	if len(report.Indicators) != 6 || counts["OUTCOME"] != 1 || counts["DRIVER"] != 2 || counts["GUARDRAIL"] != 3 || counts["FOUNDATION"] != 3 || counts["COHERENCE"] != 2 || counts["REGRESSION"] != 1 {
		t.Fatalf("counts=%v", counts)
	}
}

func TestUnknownTopNeverBecomesFixedPoint(t *testing.T) {
	report := Evaluate(CaseInput("unknown-top", candidateSHA))
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown || report.Reason != ReasonUnknown || report.Summary.UnknownPaths != 1 {
		t.Fatalf("report=%+v", report)
	}
}
