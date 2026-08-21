package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

func BuildIndicatorDecisionLedger(indicators []sourcepolicy.Indicator, actions []Action) (IndicatorDecisionLedger, error) {
	return buildIndicatorDecisionLedger(indicators, actions, nil)
}

func buildPlanIndicatorDecisionLedger(indicators []sourcepolicy.Indicator, actions []Action, deferred []string) (IndicatorDecisionLedger, error) {
	return buildIndicatorDecisionLedger(indicators, actions, deferred)
}
