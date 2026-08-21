package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func buildLedgerEntries(indicators []sourcepolicy.Indicator, actions map[string]Action) ([]IndicatorDecisionLedgerEntry, int, error) {
	entries := make([]IndicatorDecisionLedgerEntry, 0, len(indicators))
	seen := make(map[string]struct{}, len(indicators))
	selectedCount := 0
	for _, indicator := range indicators {
		id := indicatorID(indicator)
		if _, exists := seen[id]; exists {
			return nil, 0, fmt.Errorf("duplicate source indicator %q", id)
		}
		seen[id] = struct{}{}
		action, hasAction := actions[id]
		entry, selected, err := buildLedgerEntry(id, indicator, action, hasAction)
		if err != nil {
			return nil, 0, err
		}
		if selected {
			selectedCount++
		}
		entries = append(entries, entry)
	}
	if selectedCount != len(actions) {
		return nil, 0, fmt.Errorf("%d actions do not belong to the indicator set", len(actions)-selectedCount)
	}
	return entries, selectedCount, nil
}
