package generation

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestIndicatorDecisionLedgerClosesAllTrilemmaRoutes(t *testing.T) {
	exempt := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityNotApplicable,
		Satisfied:     true,
		Proof:         sourcepolicy.ProofFoundation,
	}
	conforming := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Satisfied:     true,
		Proof:         sourcepolicy.ProofCoherence,
	}
	repair := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Satisfied:     false,
		Blocking:      true,
		Proof:         sourcepolicy.ProofRegression,
	}
	action := Action{
		IndicatorID:      indicatorID(repair),
		Blocking:         repair.Blocking,
		IndicatorOutcome: repair.Outcome(),
		SourceIndicator:  repair,
	}

	ledger, err := BuildIndicatorDecisionLedger(
		[]sourcepolicy.Indicator{repair, exempt, conforming},
		[]Action{action},
	)
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

func TestIndicatorDecisionLedgerRejectsMissingRepair(t *testing.T) {
	indicator := sourcepolicy.Indicator{
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Satisfied:     false,
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
