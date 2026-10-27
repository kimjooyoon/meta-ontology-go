package prooftransport

import (
	"context"
	"strings"
	"testing"
)

func TestRealTransportRetainsSourceOwnedOutcome(t *testing.T) {
	for _, row := range []struct {
		status         int
		store          bool
		want           string
		failed, effect bool
	}{
		{201, true, "EFFECT_OBSERVED", false, true},
		{500, true, "EFFECT_OBSERVED_WITH_WRITE_FAILURE", true, true},
		{500, false, "EFFECT_NOT_OBSERVED", true, false},
	} {
		f := fixture(t, false)
		f.putStatus, f.store = row.status, row.store
		got := run(context.Background(), "transport-original-build-proof", f.material, "", true, f)
		if got.Outcome.F_state != row.want || got.Outcome.F_command_failed != row.failed || got.Outcome.F_effect_observed != row.effect || got.Outcome.F_effect_id != got.Scope.F_effect_id || !strings.Contains(got.OutcomeGenerationOrigin, "gooo://transport-outcome/assess") {
			t.Fatal(got, row)
		}
	}
}

func TestEarlyReturnAlsoAssessesReadOnlyOutcome(t *testing.T) {
	f := fixture(t, false)
	got := run(context.Background(), "transport-original-build-proof", []byte("changed"), "", true, f)
	if got.Decision.F_state != "REFUTED_MATERIAL" || got.Outcome.F_state != "NO_WRITE_ATTEMPT" || got.Outcome.F_next_operation != "restore-pinned-material" || got.Outcome.F_command_failed || f.puts != 0 {
		t.Fatal(got)
	}
}

func TestOrphanReadbackDecisionCannotBecomeObservedEffect(t *testing.T) {
	scope := Plan("transport-original-build-proof")
	decision := R_TransportStep{F_effect_id: scope.F_effect_id, F_state: "READY_TO_RESTORE"}
	readback := R_TransportStep{F_effect_id: scope.F_effect_id, F_state: "ALREADY_PRESENT"}
	report := Report{Scope: scope, Decision: decision, Write: WriteObservation{Attempted: true, Status: 201}, Next: &readback}
	report.assessOutcome()
	if report.Outcome.F_state != "INCONSISTENT_OBSERVATION" || !report.Outcome.F_command_failed || report.Outcome.F_effect_observed {
		t.Fatal(report.Outcome)
	}
}
