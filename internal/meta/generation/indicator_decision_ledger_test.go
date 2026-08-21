package generation

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestIndicatorDecisionLedgerClosesAllTrilemmaRoutes(t *testing.T) {
	indicators, actions := indicatorDecisionLedgerFixture()
	ledger, err := BuildIndicatorDecisionLedger(indicators, actions)
	if err != nil {
		t.Fatalf("BuildIndicatorDecisionLedger() error = %v", err)
	}
	if ledger.IndicatorCount != 3 || ledger.SelectedCount != 1 {
		t.Fatalf("unexpected ledger counts: %+v", ledger)
	}
	if ledger.FoundationalCount != 1 || ledger.CoherenceCount != 1 || ledger.RegressiveCount != 1 {
		t.Fatalf("unexpected trilemma counts: %+v", ledger)
	}
	if !strings.HasPrefix(ledger.Digest, "sha256:") || len(ledger.Digest) != len("sha256:")+64 {
		t.Fatalf("unexpected ledger digest %q", ledger.Digest)
	}
	payload, err := json.Marshal(ledger)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded IndicatorDecisionLedger
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(decoded, ledger) {
		t.Fatalf("round trip mismatch\n got: %+v\nwant: %+v", decoded, ledger)
	}
}
