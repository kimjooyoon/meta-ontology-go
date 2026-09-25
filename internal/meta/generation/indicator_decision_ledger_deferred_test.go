package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestPlanIndicatorDecisionLedgerRecordsDeferredRepair(t *testing.T) {
	indicators, actions := indicatorDecisionLedgerFixture()
	deferred := sourcepolicy.Indicator{
		Subject:       "deferred.go",
		Applicability: sourcepolicy.ApplicabilityApplicable,
		Blocking:      true,
		Proof:         sourcepolicy.ProofRegression,
	}
	deferredID := indicatorID(deferred)
	indicators = append(indicators, deferred)
	ledger, err := buildPlanIndicatorDecisionLedger(indicators, actions, []string{deferredID})
	if err != nil {
		t.Fatalf("buildPlanIndicatorDecisionLedger() error = %v", err)
	}
	if ledger.SelectedCount != 1 || ledger.DeferredCount != 1 {
		t.Fatalf("unexpected repair counts: %+v", ledger)
	}
	found := false
	for _, entry := range ledger.Entries {
		if entry.IndicatorID == deferredID {
			found = entry.Disposition == IndicatorDispositionRepairDeferred && entry.Action == nil
		}
	}
	if !found {
		t.Fatal("deferred repair was not represented canonically")
	}
	if err := ledger.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
