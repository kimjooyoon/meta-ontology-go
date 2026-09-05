package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func attachPlanIndicatorDecisionLedger(plan Plan, indicators []sourcepolicy.Indicator) Plan {
	if plan.Decision != DecisionFixedPoint && plan.Decision != DecisionPlan {
		return plan
	}
	ledger, err := buildPlanIndicatorDecisionLedgerWithRefuted(indicators, plan.Selected, plan.UnselectedIndicatorIDs, plan.RefutedIndicatorIDs)
	if err != nil {
		plan.Decision, plan.Reason = DecisionUnknown, ReasonInvalidInput
		plan.Selected = []Action{}
		plan.IndicatorDecisionLedger = nil
		return finish(plan)
	}
	plan.IndicatorDecisionLedger = &ledger
	return finish(plan)
}
