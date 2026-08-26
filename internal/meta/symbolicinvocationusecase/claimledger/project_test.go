package claimledger

import (
	"encoding/json"
	"testing"
)

func TestUnknownIsLocalizedAndCannotPass(t *testing.T) {
	contract := testContract([]ClaimSpec{
		{
			ID: "schema", Kind: "ASSERTION", Modality: "MUST", Subject: "observation", Predicate: "schema-matches",
			Scope: "IN_SCOPE", ProofRoute: "FOUNDATION", Coordinate: Coordinate{Stage: "OBSERVE", Step: "decode-schema"},
			Evidence: &EvidenceSpec{Paths: []string{"schema"}, Operator: "EQUALS", Expected: rawString("observed/v1")},
			UnknownReason: "SCHEMA_MISSING", RefutedReason: "SCHEMA_MISMATCH",
		},
		{
			ID: "memory", Kind: "OBLIGATION", Modality: "MUST", Subject: "execution", Predicate: "peak-rss-recorded",
			Scope: "IN_SCOPE", ProofRoute: "COHERENCE", Coordinate: Coordinate{Stage: "OBSERVE", Step: "capture-peak-rss"},
			Evidence: &EvidenceSpec{Paths: []string{"performance.peak_rss_kib"}, Operator: "POSITIVE_INTEGER"},
			UnknownReason: "PEAK_RSS_MISSING", RefutedReason: "PEAK_RSS_INVALID",
		},
		{
			ID: "human", Kind: "CANDIDATE", Modality: "MAY", Subject: "reader", Predicate: "understands-output",
			Scope: "EXCLUDED", ProofRoute: "REGRESSION", Coordinate: Coordinate{Stage: "OBSERVE", Step: "assess-human"},
			ExcludedReason: "NON_MACHINE_CLAIM_EXCLUDED",
		},
	}, ExpectedMetrics{
		FixedClaimTotal: 3, InScopeClaimTotal: 2, DischargedTotal: 1, UnknownTotal: 1,
		ExcludedTotal: 1, OpenClaimTotal: 1, DischargeBasisPoints: 5000,
		ProofRoutes: ProofRouteCounts{Foundation: 1, Coherence: 1, Regression: 1},
		ClaimSetDecision: "FAIL_CLOSED", Resolution: "STAGE_LOCAL",
	})
	report, err := Project(contract, []byte(`{"schema":"observed/v1"}`), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if report.Conformance.Decision != "PASS" || report.ClaimSet.Decision != "FAIL_CLOSED" {
		t.Fatalf("conformance=%s claim-set=%s", report.Conformance.Decision, report.ClaimSet.Decision)
	}
	if report.ClaimSet.Resolution != "STAGE_LOCAL" || report.Metrics.FalsePromotionCount != 0 {
		t.Fatalf("resolution=%s false-promotions=%d", report.ClaimSet.Resolution, report.Metrics.FalsePromotionCount)
	}
	if len(report.OpenClaimIDs) != 1 || report.OpenClaimIDs[0] != "memory" {
		t.Fatalf("open claims=%v", report.OpenClaimIDs)
	}
	if report.Claims[0].Status != "DISCHARGED" || report.Claims[1].Status != "UNKNOWN" {
		t.Fatalf("statuses=%s,%s", report.Claims[0].Status, report.Claims[1].Status)
	}
	if report.Claims[1].Coordinate.Step != "capture-peak-rss" || report.Claims[1].Reason != "PEAK_RSS_MISSING" {
		t.Fatalf("unknown coordinate=%+v reason=%s", report.Claims[1].Coordinate, report.Claims[1].Reason)
	}
}

func TestRefutedEvidenceIsConformantButNeverPromoted(t *testing.T) {
	contract := testContract([]ClaimSpec{{
		ID: "decision", Kind: "ASSERTION", Modality: "MUST", Subject: "observation", Predicate: "decision-pass",
		Scope: "IN_SCOPE", ProofRoute: "FOUNDATION", Coordinate: Coordinate{Stage: "OBSERVE", Step: "read-decision"},
		Evidence: &EvidenceSpec{Paths: []string{"decision"}, Operator: "EQUALS", Expected: rawString("PASS")},
		UnknownReason: "DECISION_MISSING", RefutedReason: "DECISION_NOT_PASS",
	}}, ExpectedMetrics{
		FixedClaimTotal: 1, InScopeClaimTotal: 1, RefutedTotal: 1,
		ProofRoutes: ProofRouteCounts{Foundation: 1},
		ClaimSetDecision: "FAIL_CLOSED", Resolution: "CLAIM_LOCAL",
	})
	report, err := Project(contract, []byte(`{"decision":"UNKNOWN"}`), "abc")
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

func TestClaimRequiresAStageAndStep(t *testing.T) {
	contract := testContract([]ClaimSpec{{
		ID: "missing-step", Kind: "OBLIGATION", Modality: "MUST", Subject: "artifact", Predicate: "exists",
		Scope: "IN_SCOPE", ProofRoute: "FOUNDATION", Coordinate: Coordinate{Stage: "EMIT"},
		Evidence: &EvidenceSpec{Paths: []string{"artifact"}, Operator: "NON_NULL"},
		UnknownReason: "ARTIFACT_MISSING", RefutedReason: "ARTIFACT_INVALID",
	}}, ExpectedMetrics{FixedClaimTotal: 1})
	if _, err := Project(contract, []byte(`{}`), "abc"); err == nil {
		t.Fatal("expected an imprecise process coordinate to fail")
	}
}

func testContract(claims []ClaimSpec, expected ExpectedMetrics) []byte {
	encoded, err := json.Marshal(Contract{
		Schema: ContractSchema, Metric: "gooo.metric.test.claim-ledger.v1", Expected: expected, Claims: claims,
	})
	if err != nil {
		panic(err)
	}
	return encoded
}

func rawString(value string) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}
