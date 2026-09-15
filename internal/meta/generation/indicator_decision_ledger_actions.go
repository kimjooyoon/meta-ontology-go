package generation

import (
	"fmt"
	"reflect"
)

func indexLedgerActions(actions []Action) (map[string]Action, error) {
	indexed := make(map[string]Action, len(actions))
	for _, action := range actions {
		if !actionMatchesSourceIndicator(action) {
			return nil, fmt.Errorf("action %q does not match its source indicator", action.IndicatorID)
		}
		if !reflect.DeepEqual(action.IndicatorOutcome, action.SourceIndicator.Outcome()) {
			return nil, fmt.Errorf("action %q carries a forged indicator outcome", action.IndicatorID)
		}
		if _, exists := indexed[action.IndicatorID]; exists {
			return nil, fmt.Errorf("duplicate action for indicator %q", action.IndicatorID)
		}
		indexed[action.IndicatorID] = action
	}
	return indexed, nil
}
