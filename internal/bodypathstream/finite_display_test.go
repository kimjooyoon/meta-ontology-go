package bodypathstream

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestFiniteExpectationDisplayUsesObservedCasesAndReplay(t *testing.T) {
	// These synthetic observations cover display states, not native execution.
	for _, tc := range []struct {
		name     string
		observed *bodyexecution.Observation
		passed   int
		want     string
	}{
		{name: "rejected before execution", want: "unobserved"},
		{name: "unknown count", observed: &bodyexecution.Observation{}, want: "unobserved"},
		{name: "declared before execution", observed: &bodyexecution.Observation{DeclaredCases: 128},
			want: "unobserved (128 declared)"},
		{name: "started without decoded outputs", observed: &bodyexecution.Observation{DeclaredCases: 2,
			Runs: []bodyexecution.ProcessObservation{{Started: true}}}, want: "unobserved (2 declared)"},
		{name: "observed zero", observed: &bodyexecution.Observation{DeclaredCases: 2, RuntimeReplayed: true,
			Cases: make([]bodycodegen.IRBodyFillCaseResult, 2)}, want: "0/2"},
		{name: "completed replay", observed: &bodyexecution.Observation{DeclaredCases: 2, RuntimeReplayed: true,
			Cases: []bodycodegen.IRBodyFillCaseResult{{Passed: true}, {}}}, passed: 1, want: "1/2"},
		{name: "first run before failed replay", observed: &bodyexecution.Observation{DeclaredCases: 2,
			Cases: []bodycodegen.IRBodyFillCaseResult{{Passed: true}, {}}}, passed: 1,
			want: "observed 1/2 (replay incomplete)"},
		{name: "partial observations", observed: &bodyexecution.Observation{DeclaredCases: 2,
			Cases: []bodycodegen.IRBodyFillCaseResult{{Passed: true}}}, passed: 1,
			want: "observed 1/1 (2 declared; replay incomplete)"},
		{name: "replayed partial count", observed: &bodyexecution.Observation{DeclaredCases: 2, RuntimeReplayed: true,
			Cases: []bodycodegen.IRBodyFillCaseResult{{Passed: true}}}, passed: 1,
			want: "observed 1/1 (2 declared)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result Result
			if tc.observed != nil {
				result.Execution = &bodyexecution.Result{Observation: *tc.observed}
			}
			if got := finiteExpectationDisplay(result, tc.passed); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
