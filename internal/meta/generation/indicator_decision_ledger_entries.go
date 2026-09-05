package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func buildLedgerEntries(indicators []sourcepolicy.Indicator, actions map[string]Action, deferred map[string]struct{}) ([]IndicatorDecisionLedgerEntry, int, int, error) {
	entries, selectedCount, deferredCount, _, err := buildLedgerEntriesWithRefuted(indicators, actions, deferred, nil)
	return entries, selectedCount, deferredCount, err
}

func buildLedgerEntriesWithRefuted(indicators []sourcepolicy.Indicator, actions map[string]Action, deferred, refuted map[string]struct{}) ([]IndicatorDecisionLedgerEntry, int, int, int, error) {
	entries := make([]IndicatorDecisionLedgerEntry, 0, len(indicators))
	seen := make(map[string]struct{}, len(indicators))
	selectedCount := 0
	deferredCount := 0
	refutedCount := 0
	for _, indicator := range indicators {
		id := indicatorID(indicator)
		if _, exists := seen[id]; exists {
			return nil, 0, 0, fmt.Errorf("duplicate source indicator %q", id)
		}
		seen[id] = struct{}{}
		action, hasAction := actions[id]
		_, isDeferred := deferred[id]
		_, isRefuted := refuted[id]
		entry, err := buildLedgerEntryWithRefuted(id, indicator, action, hasAction, isDeferred, isRefuted)
		if err != nil {
			return nil, 0, 0, 0, err
		}
		if entry.Action != nil {
			selectedCount++
		}
		if entry.Disposition == IndicatorDispositionRepairDeferred {
			deferredCount++
		}
		if entry.Disposition == IndicatorDispositionRepairRefuted {
			refutedCount++
		}
		entries = append(entries, entry)
	}
	if selectedCount != len(actions) {
		return nil, 0, 0, 0, fmt.Errorf("%d actions do not belong to the indicator set", len(actions)-selectedCount)
	}
	if deferredCount != len(deferred) {
		return nil, 0, 0, 0, fmt.Errorf("%d deferred repairs do not belong to the indicator set", len(deferred)-deferredCount)
	}
	if refutedCount != len(refuted) {
		return nil, 0, 0, 0, fmt.Errorf("%d refuted repairs do not belong to the indicator set", len(refuted)-refutedCount)
	}
	return entries, selectedCount, deferredCount, refutedCount, nil
}
