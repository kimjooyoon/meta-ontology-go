package cyclecontext_test

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/cyclecontext"
)

func TestSourceCycleLifecycleExpectations(t *testing.T) {
	for _, row := range []struct {
		conclusion, status, reason string
		supported                  bool
	}{
		{"success", "BOUND", "SELF_IMPROVEMENT_CYCLE_BOUND", true},
		{"failure", "BOUND", "SELF_IMPROVEMENT_CYCLE_BOUND", true},
		{"cancelled", "INTERRUPTED", "SOURCE_CI_CANCELLED", true},
		{"", "OPEN", "SELF_IMPROVEMENT_CYCLE_OPEN", false},
		{"SUCCESS", "OPEN", "SELF_IMPROVEMENT_CYCLE_OPEN", false},
		{"cancelled ", "OPEN", "SELF_IMPROVEMENT_CYCLE_OPEN", false},
		{"timed_out", "OPEN", "SELF_IMPROVEMENT_CYCLE_OPEN", false},
	} {
		for _, valid := range []bool{false, true} {
			got := cyclecontext.Evaluate(cyclecontext.R_Observation{F_conclusion: row.conclusion, F_bindings_valid: valid})
			status, reason := row.status, row.reason
			if !valid {
				status, reason = "OPEN", "SELF_IMPROVEMENT_CYCLE_OPEN"
			}
			if got.F_status != status || got.F_reason != reason || got.F_conclusion_supported != row.supported ||
				got.F_record_valid != (valid && row.supported) {
				t.Fatalf("conclusion=%q bindings=%v got=%+v", row.conclusion, valid, got)
			}
		}
	}
}
