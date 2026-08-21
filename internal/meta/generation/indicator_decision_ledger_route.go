package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func buildApplicableLedgerEntry(entry IndicatorDecisionLedgerEntry, action Action, hasAction, deferred bool) (IndicatorDecisionLedgerEntry, error) {
	if entry.SourceIndicator.Satisfied {
		if hasAction || deferred {
			return IndicatorDecisionLedgerEntry{}, fmt.Errorf("conforming indicator %q selected a repair", entry.IndicatorID)
		}
		entry.Disposition = IndicatorDispositionConforming
		return entry, nil
	}
	if hasAction && deferred {
		return IndicatorDecisionLedgerEntry{}, fmt.Errorf("violating indicator %q is both selected and deferred", entry.IndicatorID)
	}
	if hasAction {
		actionCopy := action
		entry.Disposition = IndicatorDispositionRepairSelected
		entry.Action = &actionCopy
		return entry, nil
	}
	if deferred {
		entry.Disposition = IndicatorDispositionRepairDeferred
		return entry, nil
	}
	return IndicatorDecisionLedgerEntry{}, fmt.Errorf("violating indicator %q has no selected repair", entry.IndicatorID)
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
