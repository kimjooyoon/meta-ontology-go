package generation

import (
	"encoding/json"
	"testing"
)

func TestPlanProjectsIndicatorDecisionSummary(t *testing.T) {
	plan := Plan{
		IndicatorsDigest:          "sha256:source",
		Selected:                  make([]Action, 2),
		UnselectedIndicatorIDs:    []string{"unselected"},
		NotApplicableIndicatorIDs: []string{"root-a", "root-b"},
		UnknownIndicatorIDs:       []string{"unknown"},
	}
	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Summary IndicatorDecisionSummary `json:"indicator_decision_summary"`
	}
	if err := json.Unmarshal(payload, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.Summary.IndicatorsDigest != plan.IndicatorsDigest ||
		wire.Summary.SelectedFailClosedCount != 2 ||
		wire.Summary.NotApplicableCount != 2 ||
		wire.Summary.UnknownCount != 1 {
		t.Fatalf("summary = %+v", wire.Summary)
	}
}

func TestExecutionProjectsPlanDecisionSummary(t *testing.T) {
	manifest := ExecutionManifest{
		PlanDigest:                "sha256:plan",
		Steps:                     make([]ExecutionStep, 3),
		NotApplicableIndicatorIDs: []string{"root-a", "root-b"},
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Summary ExecutionDecisionSummary `json:"indicator_decision_summary"`
	}
	if err := json.Unmarshal(payload, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.Summary.PlanDigest != manifest.PlanDigest ||
		wire.Summary.ExecutableFailClosedCount != 3 ||
		wire.Summary.NotApplicableIndicatorCount != 2 {
		t.Fatalf("summary = %+v", wire.Summary)
	}
}
