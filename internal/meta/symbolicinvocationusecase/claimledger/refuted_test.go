package claimledger

import "testing"

func TestRefutedEvidenceIsConformantButNeverPromoted(t *testing.T) {
	claims := []ClaimSpec{
		inScope("decision", "decision", "EQUALS", "FOUNDATION", "OBSERVE", "read-decision", rawString("PASS")),
	}
	expected := ExpectedMetrics{
		FixedClaimTotal: 1, InScopeClaimTotal: 1, RefutedTotal: 1,
		ProofRoutes: ProofRouteCounts{Foundation: 1},
		ClaimSetDecision: "FAIL_CLOSED", Resolution: "CLAIM_LOCAL",
	}
	report, err := Project(testContract(claims, expected), []byte(`{"decision":"UNKNOWN"}`), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if report.Conformance.Decision != "PASS" || report.Claims[0].Status != "REFUTED" {
		t.Fatalf("conformance=%s status=%s", report.Conformance.Decision, report.Claims[0].Status)
	}
	if report.Metrics.FalsePromotionCount != 0 || report.ClaimSet.Decision == "PASS" {
		t.Fatalf("false-promotions=%d decision=%s", report.Metrics.FalsePromotionCount, report.ClaimSet.Decision)
	}
}
