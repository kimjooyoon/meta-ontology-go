package sourcepolicy

import (
	"encoding/json"
	"testing"
)

func TestProjectRootTopologyHasClosedNotApplicableDecision(t *testing.T) {
	indicator := Indicator{
		MetricID:      "gooo.metric.layout.direct-entries.v1",
		Applicability: "NOT_APPLICABLE",
		Satisfied:     true,
	}
	outcome := indicator.Outcome()
	if outcome.Decision != IndicatorDecisionNotApplicable {
		t.Fatalf("decision = %q", outcome.Decision)
	}
	if outcome.FailureReason != FailureReasonCatalogNotApplicable {
		t.Fatalf("failure reason = %q", outcome.FailureReason)
	}
	if outcome.EnforcementEffect != EnforcementEffectNone {
		t.Fatalf("enforcement effect = %q", outcome.EnforcementEffect)
	}
	payload, err := json.Marshal(indicator)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Decision          IndicatorDecision `json:"decision"`
		EnforcementEffect EnforcementEffect `json:"enforcement_effect"`
	}
	if err := json.Unmarshal(payload, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.Decision != outcome.Decision ||
		wire.EnforcementEffect != outcome.EnforcementEffect {
		t.Fatalf("wire outcome = %+v", wire)
	}
}
