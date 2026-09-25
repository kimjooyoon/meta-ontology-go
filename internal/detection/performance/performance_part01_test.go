package performance

import (
	"strings"
	"testing"
)

func TestMeasureAllUsesDeterministicOperationCounts(t *testing.T) {
	specs := syntheticSpecs()
	config := Config{Iterations: 4, AllocationRuns: 2}
	report, err := MeasureAll(specs, config)
	if err != nil {
		t.Fatalf("MeasureAll() error = %v", err)
	}
	if !report.Passed() {
		t.Fatalf("MeasureAll() unexpected violations = %#v", report.Violations)
	}
	if got := report.Observations[0].Operations; got != 12 {
		t.Fatalf("parser operations = %d, want 12", got)
	}
	if got := report.Observations[3].Operations; got != 28 {
		t.Fatalf("generator operations = %d, want 28", got)
	}
	if !strings.Contains(report.String(), "status=pass") {
		t.Fatalf("report did not contain pass status:\n%s", report)
	}
}
func TestMeasureAllCanonicalizesPermutationsWithoutMutation(t *testing.T) {
	forward := syntheticSpecs()
	permuted := []StageSpec{forward[4], forward[1], forward[3], forward[0], forward[2]}
	wantInput := append([]StageSpec(nil), permuted...)
	config := Config{Iterations: 4, AllocationRuns: 2}
	want, err := MeasureAll(forward, config)
	if err != nil {
		t.Fatalf("MeasureAll(forward) error = %v", err)
	}
	got, err := MeasureAll(permuted, config)
	if err != nil {
		t.Fatalf("MeasureAll(permuted) error = %v", err)
	}
	for i := range permuted {
		if permuted[i].Stage != wantInput[i].Stage || permuted[i].Budget != wantInput[i].Budget {
			t.Fatalf("MeasureAll mutated input spec %d: got %#v want %#v", i, permuted[i], wantInput[i])
		}
	}
	if len(got.Observations) != len(StandardStages()) {
		t.Fatalf("observation count = %d, want %d", len(got.Observations), len(StandardStages()))
	}
	if got.String() != want.String() {
		t.Fatalf("permuted report differs:\n got %s\nwant %s", got, want)
	}
	for i, stage := range StandardStages() {
		if got.Observations[i].Stage != stage {
			t.Fatalf("observation %d stage = %q, want %q", i, got.Observations[i].Stage, stage)
		}
	}
}
