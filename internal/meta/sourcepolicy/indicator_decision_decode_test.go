package sourcepolicy

import (
	"encoding/json"
	"testing"
)

func TestIndicatorOutcomeRoundTripsStrictly(t *testing.T) {
	original := Indicator{
		MetricID:      "gooo.metric.layout.direct-entries.v1",
		Applicability: "NOT_APPLICABLE", Satisfied: true,
	}
	payload, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Indicator
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != original {
		t.Fatalf("decoded indicator = %+v", decoded)
	}
}

func TestIndicatorRejectsForgedOutcome(t *testing.T) {
	indicator := Indicator{
		MetricID:      "gooo.metric.source.file-lines.v1",
		Applicability: "APPLICABLE", Blocking: true,
	}
	payload, err := json.Marshal(indicator)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(payload, &wire); err != nil {
		t.Fatal(err)
	}
	wire["decision"] = string(IndicatorDecisionPass)
	forged, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Indicator
	if err := json.Unmarshal(forged, &decoded); err == nil {
		t.Fatal("forged indicator outcome accepted")
	}
}
