package generation

import "fmt"

func indexDeferredIndicatorIDs(ids []string) (map[string]struct{}, error) {
	indexed := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if !validIndicatorDecisionLedgerDigest(id) {
			return nil, fmt.Errorf("invalid deferred indicator id %q", id)
		}
		if _, exists := indexed[id]; exists {
			return nil, fmt.Errorf("duplicate deferred indicator id %q", id)
		}
		indexed[id] = struct{}{}
	}
	return indexed, nil
}
