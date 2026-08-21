package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func buildApplicableLedgerEntry(entry IndicatorDecisionLedgerEntry, action Action, hasAction bool) (IndicatorDecisionLedgerEntry, bool, error) {
	if entry.SourceIndicator.Satisfied {
		if hasAction {
			return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("conforming indicator %q selected an action", entry.IndicatorID)
		}
		entry.Disposition = IndicatorDispositionConforming
		return entry, false, nil
	}
	if !hasAction {
		return IndicatorDecisionLedgerEntry{}, false, fmt.Errorf("violating indicator %q has no selected repair", entry.IndicatorID)
	}
	actionCopy := action
	entry.Disposition = IndicatorDispositionRepairSelected
	entry.Action = &actionCopy
	return entry, true, nil
}

func indicatorTrilemmaRoute(proof sourcepolicy.ProofChoice) (TrilemmaRoute, error) {
	switch proof {
	case sourcepolicy.ProofFoundation:
		return TrilemmaRouteFoundation, nil
	case sourcepolicy.ProofCoherence:
		return TrilemmaRouteCoherence, nil
	case sourcepolicy.ProofRegression:
		return TrilemmaRouteRegression, nil
	default:
		return "", fmt.Errorf("unknown proof choice %q", proof)
	}
}
