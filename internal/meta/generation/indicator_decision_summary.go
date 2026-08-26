package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

type IndicatorDecisionSummary struct {
	SourceSchema              string `json:"source_schema"`
	IndicatorsDigest          string `json:"indicators_digest"`
	SelectedFailClosedCount   int    `json:"selected_fail_closed_count"`
	UnselectedFailClosedCount int    `json:"unselected_fail_closed_count"`
	NotApplicableCount        int    `json:"not_applicable_count"`
	UnknownCount              int    `json:"unknown_count"`
}

func (plan Plan) indicatorDecisionSummary() IndicatorDecisionSummary {
	return IndicatorDecisionSummary{
		SourceSchema:              sourcepolicy.IndicatorSchema,
		IndicatorsDigest:          plan.IndicatorsDigest,
		SelectedFailClosedCount:   len(plan.Selected),
		UnselectedFailClosedCount: len(plan.UnselectedIndicatorIDs),
		NotApplicableCount:        len(plan.NotApplicableIndicatorIDs),
		UnknownCount:              len(plan.UnknownIndicatorIDs),
	}
}

type ExecutionDecisionSummary struct {
	PlanDigest                  string `json:"plan_digest"`
	ExecutableFailClosedCount   int    `json:"executable_fail_closed_count"`
	NotApplicableIndicatorCount int    `json:"not_applicable_indicator_count"`
}

func (manifest ExecutionManifest) indicatorDecisionSummary() ExecutionDecisionSummary {
	return ExecutionDecisionSummary{
		PlanDigest:                  manifest.PlanDigest,
		ExecutableFailClosedCount:   len(manifest.Steps),
		NotApplicableIndicatorCount: len(manifest.NotApplicableIndicatorIDs),
	}
}
