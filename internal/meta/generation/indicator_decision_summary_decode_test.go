package generation

import (
	"encoding/json"
	"testing"
)

func TestPlanDecisionSummaryRoundTripsStrictly(t *testing.T) {
	original := Plan{
		IndicatorsDigest:          "sha256:source",
		NotApplicableIndicatorIDs: []string{"root"},
	}
	payload, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Plan
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.IndicatorsDigest != original.IndicatorsDigest {
		t.Fatalf("decoded plan = %+v", decoded)
	}
}

func TestPlanRejectsForgedDecisionSummary(t *testing.T) {
	plan := Plan{IndicatorsDigest: "sha256:source"}
	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(payload, &wire); err != nil {
		t.Fatal(err)
	}
	summary := wire["indicator_decision_summary"].(map[string]any)
	summary["not_applicable_count"] = float64(99)
	forged, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Plan
	if err := json.Unmarshal(forged, &decoded); err == nil {
		t.Fatal("forged decision summary accepted")
	}
}
