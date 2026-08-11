package performance

import (
	"errors"
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

func TestRejectsDuplicateStages(t *testing.T) {
	spec := StageSpec{Stage: ParserStage, Run: func(*Counter) error { return nil }}
	_, err := MeasureAll([]StageSpec{spec, spec}, Config{Iterations: 1, AllocationRuns: 1})
	if err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("MeasureAll() error = %v, want duplicate-stage error", err)
	}
}

func syntheticSpecs() []StageSpec {
	return []StageSpec{
		{Stage: ParserStage, Run: fixedWork(3), Budget: Budget{MaxOperationsPerIteration: 3}},
		{Stage: SemanticStage, Run: fixedWork(5), Budget: Budget{MaxOperationsPerIteration: 5}},
		{Stage: QueryStage, Run: fixedWork(2), Budget: Budget{MaxOperationsPerIteration: 2}},
		{Stage: GeneratorStage, Run: fixedWork(7), Budget: Budget{MaxOperationsPerIteration: 7}},
		{Stage: CacheStage, Run: fixedWork(1), Budget: Budget{MaxOperationsPerIteration: 1}},
	}
}

func fixedWork(operations uint64) Runner {
	return func(counter *Counter) error {
		counter.Add(operations)
		return nil
	}
}

func BenchmarkSyntheticPipeline(b *testing.B) {
	for _, spec := range syntheticSpecs() {
		b.Run(string(spec.Stage), func(b *testing.B) {
			BenchmarkStage(b, spec)
		})
	}
}
