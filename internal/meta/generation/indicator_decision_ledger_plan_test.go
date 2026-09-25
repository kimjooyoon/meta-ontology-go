package generation

import (
	"encoding/json"
	"testing"
)

func TestPlanJSONReplaysIndicatorDecisionLedger(t *testing.T) {
	indicators, actions := indicatorDecisionLedgerFixture()
	normalized := normalizeIndicators(indicators)
	plan := Plan{
		Decision:                  DecisionPlan,
		Reason:                    ReasonIndependentActions,
		IndicatorsDigest:          digestJSON(normalized),
		NotApplicableIndicatorIDs: notApplicableIndicatorIDs(normalized),
		Selected:                  actions,
	}
	plan = attachPlanIndicatorDecisionLedger(plan, indicators)
	if plan.IndicatorDecisionLedger == nil {
		t.Fatal("plan has no indicator decision ledger")
	}
	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded Plan
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	decoded.IndicatorsDigest = "sha256:" + ledgerZeroDigest
	payload, err = json.Marshal(decoded)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := json.Unmarshal(payload, &Plan{}); err == nil {
		t.Fatal("json.Unmarshal() accepted a forged indicator digest")
	}
}
