package generation

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func validatePlanIndicatorDecisionLedger(plan Plan) error {
	required := plan.Decision == DecisionFixedPoint || plan.Decision == DecisionPlan
	if !required {
		if plan.IndicatorDecisionLedger != nil {
			return fmt.Errorf("non-executable plan carries an indicator decision ledger")
		}
		return nil
	}
	if plan.IndicatorDecisionLedger == nil {
		return fmt.Errorf("executable plan has no indicator decision ledger")
	}
	if err := plan.IndicatorDecisionLedger.Validate(); err != nil {
		return err
	}
	indicators := ledgerSourceIndicators(*plan.IndicatorDecisionLedger)
	expected, err := buildPlanIndicatorDecisionLedgerWithRefuted(indicators, plan.Selected, plan.UnselectedIndicatorIDs, plan.RefutedIndicatorIDs)
	if err != nil || !reflect.DeepEqual(*plan.IndicatorDecisionLedger, expected) {
		return fmt.Errorf("plan indicator decision ledger does not match plan decisions")
	}
	normalized := normalizeIndicators(indicators)
	notApplicable := notApplicableIndicatorIDs(normalized)
	sort.Strings(notApplicable)
	if plan.IndicatorsDigest != digestJSON(normalized) ||
		!reflect.DeepEqual(plan.NotApplicableIndicatorIDs, notApplicable) {
		return fmt.Errorf("plan indicator decision ledger does not match source indicators")
	}
	return nil
}

func ledgerSourceIndicators(ledger IndicatorDecisionLedger) []sourcepolicy.Indicator {
	indicators := make([]sourcepolicy.Indicator, 0, len(ledger.Entries))
	for _, entry := range ledger.Entries {
		indicators = append(indicators, entry.SourceIndicator)
	}
	return indicators
}
