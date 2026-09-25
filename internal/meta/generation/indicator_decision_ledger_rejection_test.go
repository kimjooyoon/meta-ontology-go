package generation

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestIndicatorDecisionLedgerRejectsMissingRepair(t *testing.T) {
	indicator := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Blocking:      true,
		Proof:         sourcepolicy.ProofRegression,
	}
	_, err := BuildIndicatorDecisionLedger([]sourcepolicy.Indicator{indicator}, nil)
	if err == nil || !strings.Contains(err.Error(), "no selected repair") {
		t.Fatalf("BuildIndicatorDecisionLedger() error = %v, want missing repair", err)
	}
}

func TestIndicatorDecisionLedgerRejectsForgedDigest(t *testing.T) {
	indicator := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Satisfied:     true,
		Proof:         sourcepolicy.ProofCoherence,
	}
	ledger, err := BuildIndicatorDecisionLedger([]sourcepolicy.Indicator{indicator}, nil)
	if err != nil {
		t.Fatalf("BuildIndicatorDecisionLedger() error = %v", err)
	}
	ledger.Digest = "sha256:" + strings.Repeat("0", 64)
	payload, err := json.Marshal(ledger)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var decoded IndicatorDecisionLedger
	if err := json.Unmarshal(payload, &decoded); err == nil {
		t.Fatal("json.Unmarshal() accepted a forged ledger digest")
	}
}
