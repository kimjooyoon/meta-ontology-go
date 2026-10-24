package generation

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func validPlanPrerequisites(plan Plan) bool {
	for _, action := range plan.Selected {
		if NativeExecutionRoute(action.Operation) == "" {
			return false
		}
	}
	if len(plan.Prerequisites) == 0 {
		return true
	}
	if plan.IndicatorDecisionLedger == nil {
		return false
	}
	indicators, err := refutedSourceIndicators(plan)
	if err != nil {
		return false
	}
	observations := make(map[string]CandidatePrerequisite, len(plan.Prerequisites))
	for index, observation := range plan.Prerequisites {
		if index > 0 && plan.Prerequisites[index-1].IndicatorID >= observation.IndicatorID {
			return false
		}
		indicator, exists := indicators[observation.IndicatorID]
		binding, known := BindingForOperation(plan.Registry, indicator.Operation)
		if !exists || !known || indicator.Satisfied || indicator.Applicability != sourcepolicy.ApplicabilityApplicable {
			return false
		}
		action := actionFor(candidate{indicator: indicator, binding: binding}, observation.IndicatorID)
		if !validCandidatePrerequisite(observation, action, plan.HeadSHA) ||
			observation.MetricsDigest != plan.Prerequisites[0].MetricsDigest {
			return false
		}
		if observation.State != "READY" && !slices.Contains(plan.UnselectedIndicatorIDs, observation.IndicatorID) {
			return false
		}
		observations[observation.IndicatorID] = observation
	}
	for _, action := range plan.Selected {
		if requiresSourcePrerequisite(action.Operation) && observations[action.IndicatorID].State != "READY" {
			return false
		}
	}
	return true
}
