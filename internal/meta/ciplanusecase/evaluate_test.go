package ciplanusecase

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

func TestFixedScorecardCountsOnlyConformedCases(t *testing.T) {
	input := completeInput(t)
	report := Evaluate(input)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.CasesSatisfied != 12 || report.Summary.PassDecisions != 4 || report.Summary.FailClosedDecisions != 4 || report.Summary.UnknownDecisions != 4 {
		t.Fatalf("decision denominator drifted: %+v", report.Summary)
	}
	if report.Summary.PersistentClaims != 36 || report.Summary.DirectUnknownClaims != 4 || report.Summary.DependencyBlocked != 8 || report.Summary.RefutedClaims != 4 {
		t.Fatalf("claim coordinates drifted: %+v", report.Summary)
	}
	if report.Summary.RuleEvidenceRefs != 6 || report.Summary.GoldenPlans != 4 || report.Summary.DeterministicReplays != 12 {
		t.Fatalf("evidence denominator drifted: %+v", report.Summary)
	}
	broken := input.Reports["pass-go"]
	broken.Claims[2].Status = metainvocation.ClaimOpen
	input.Reports["pass-go"] = broken
	degraded := Evaluate(input)
	if degraded.Summary.CasesSatisfied != 11 || degraded.Decision != "FAIL_CLOSED" {
		t.Fatalf("non-conforming decision was counted as satisfied: %+v", degraded.Summary)
	}
}
