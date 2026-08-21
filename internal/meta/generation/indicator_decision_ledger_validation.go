package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func (ledger IndicatorDecisionLedger) Validate() error {
	indicators := make([]sourcepolicy.Indicator, 0, len(ledger.Entries))
	actions := make([]Action, 0, ledger.SelectedCount)
	for _, entry := range ledger.Entries {
		indicators = append(indicators, entry.SourceIndicator)
		if entry.Action != nil {
			actions = append(actions, *entry.Action)
		}
	}
	rebuilt, err := BuildIndicatorDecisionLedger(indicators, actions)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(ledger, rebuilt) {
		return fmt.Errorf("indicator decision ledger does not match its canonical replay")
	}
	return nil
}

func (ledger *IndicatorDecisionLedger) UnmarshalJSON(data []byte) error {
	type wire IndicatorDecisionLedger
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var candidate wire
	if err := decoder.Decode(&candidate); err != nil {
		return err
	}
	if err := ensureIndicatorLedgerEOF(decoder); err != nil {
		return err
	}
	decoded := IndicatorDecisionLedger(candidate)
	if err := decoded.Validate(); err != nil {
		return err
	}
	*ledger = decoded
	return nil
}
