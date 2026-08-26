package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func buildLedgerEntry(id string, indicator sourcepolicy.Indicator, action Action, hasAction, deferred bool) (IndicatorDecisionLedgerEntry, error) {
	route, err := indicatorTrilemmaRoute(indicator.Proof)
	if err != nil {
		return IndicatorDecisionLedgerEntry{}, fmt.Errorf("indicator %q: %w", id, err)
	}
	entry := IndicatorDecisionLedgerEntry{IndicatorID: id, SourceIndicator: indicator,
		IndicatorOutcome: indicator.Outcome(), TrilemmaRoute: route}
	switch indicator.Applicability {
	case sourcepolicy.ApplicabilityNotApplicable:
		if !indicator.Satisfied || deferred {
			return IndicatorDecisionLedgerEntry{}, fmt.Errorf("not-applicable indicator %q must be closed", id)
		}
		if hasAction {
			return IndicatorDecisionLedgerEntry{}, fmt.Errorf("not-applicable indicator %q selected an action", id)
		}
		entry.Disposition = IndicatorDispositionExempt
		return entry, nil
	case sourcepolicy.ApplicabilityApplicable:
		return buildApplicableLedgerEntry(entry, action, hasAction, deferred)
	default:
		return IndicatorDecisionLedgerEntry{}, fmt.Errorf("indicator %q has unknown applicability %q", id, indicator.Applicability)
	}
}
