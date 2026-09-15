package performance

import (
	"errors"
	"testing"
)

func TestDetectsOperationAndAllocationBudgetOverruns(t *testing.T) {
	observations := []Observation{
		{
			Stage:      ParserStage,
			Iterations: 2,
			Operations: 10,
			Budget:     Budget{MaxOperationsPerIteration: 4},
		},
		{
			Stage:                   CacheStage,
			Iterations:              1,
			AllocationsPerIteration: 3,
			Budget:                  Budget{MaxAllocsPerIteration: 2},
		},
	}
	var violations []Violation
	for _, observation := range observations {
		violations = append(violations, DetectBudget(observation)...)
	}
	if len(violations) != 2 {
		t.Fatalf("violations = %#v, want two entries", violations)
	}
	if violations[0].Metric != OperationsMetric || violations[1].Metric != AllocsMetric {
		t.Fatalf("violation metrics = %#v", violations)
	}
}
func TestMeasurePropagatesRunnerErrors(t *testing.T) {
	want := errors.New("fixture failed")
	_, err := Measure(StageSpec{Stage: QueryStage, Run: func(*Counter) error {
		return want
	}}, Config{Iterations: 1, AllocationRuns: 1})
	if !errors.Is(err, want) {
		t.Fatalf("Measure() error = %v, want %v", err, want)
	}
}
