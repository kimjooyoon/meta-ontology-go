package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func buildLedgerEntry(id string, indicator sourcepolicy.Indicator, action Action, hasAction bool) (IndicatorDecisionLedgerEntry, bool, error) {
	route, err := indicatorTrilemmaRoute(indicator.Proof)
	if err != nil {
		return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("indicator %q: %w", id, err)
	}
	entry := IndicatorDecisionLedgerEntry{IndicatorID: id, SourceIndicator: indicator,
		IndicatorOutcome: indicator.Outcome(), TrilemmaRoute: route}
	switch indicator.Applicability {
	case sourcepolicy.ApplicabilityNotApplicable:
		if !indicator.Satisfied {
			return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("not-applicable indicator %q must be closed", id)
		}
		if hasAction {
			return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("not-applicable indicator %q selected an action", id)
		}
		entry.Disposition = IndicatorDispositionExempt
		return entry, false, nil
	case sourcepolicy.ApplicabilityApplicable:
		return buildApplicableLedgerEntry(entry, action, hasAction)
	default:
		return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("indicator %q has unknown applicability %q", id, indicator.Applicability)
	}
}
