package generation

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func buildIndicatorDecisionLedger(indicators []sourcepolicy.Indicator, actions []Action, deferredIDs []string) (IndicatorDecisionLedger, error) {
	actionsByIndicator, err := indexLedgerActions(actions)
	if err != nil {
		return IndicatorDecisionLedger{}, err
	}
	deferred, err := indexDeferredIndicatorIDs(deferredIDs)
	if err != nil {
		return IndicatorDecisionLedger{}, err
	}
	entries, selectedCount, deferredCount, err := buildLedgerEntries(indicators, actionsByIndicator, deferred)
	if err != nil {
		return IndicatorDecisionLedger{}, err
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].IndicatorID < entries[right].IndicatorID
	})
	ledger := IndicatorDecisionLedger{
		SchemaVersion:  IndicatorDecisionLedgerSchemaVersion,
		IndicatorCount: len(entries),
		SelectedCount:  selectedCount,
		DeferredCount:  deferredCount,
		Entries:        entries,
	}
	for _, entry := range entries {
		switch entry.TrilemmaRoute {
		case TrilemmaRouteFoundation:
			ledger.FoundationalCount++
		case TrilemmaRouteCoherence:
			ledger.CoherenceCount++
		case TrilemmaRouteRegression:
			ledger.RegressiveCount++
		}
	}
	digest, err := indicatorDecisionLedgerDigest(ledger)
	if err != nil {
		return IndicatorDecisionLedger{}, err
	}
	ledger.Digest = digest
	return ledger, nil
}
