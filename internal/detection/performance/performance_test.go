package performance

import (
	"errors"
	"reflect"
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

func TestRejectsMalformedStageAndConfig(t *testing.T) {
	valid := StageSpec{Stage: ParserStage, Run: fixedWork(1)}
	tests := []struct {
		name   string
		spec   StageSpec
		config Config
		want   string
	}{
		{
			name: "empty stage", spec: StageSpec{Run: valid.Run},
			config: Config{Iterations: 1, AllocationRuns: 1}, want: "stage is empty",
		},
		{
			name: "unknown stage", spec: StageSpec{Stage: Stage("optimizer"), Run: valid.Run},
			config: Config{Iterations: 1, AllocationRuns: 1}, want: "not a standard compiler stage",
		},
		{
			name: "missing runner", spec: StageSpec{Stage: ParserStage},
			config: Config{Iterations: 1, AllocationRuns: 1}, want: "has no runner",
		},
		{
			name: "negative iterations", spec: valid,
			config: Config{Iterations: -1, AllocationRuns: 1}, want: "iterations must be positive",
		},
		{
			name: "negative allocation runs", spec: valid,
			config: Config{Iterations: 1, AllocationRuns: -1}, want: "allocation runs must be positive",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Measure(test.spec, test.config)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Measure() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestRejectsDuplicateStages(t *testing.T) {
	spec := StageSpec{Stage: ParserStage, Run: func(*Counter) error { return nil }}
	_, err := MeasureAll([]StageSpec{spec, spec}, Config{Iterations: 1, AllocationRuns: 1})
	if err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("MeasureAll() error = %v, want duplicate-stage error", err)
	}
}

func TestStandardStagesReturnsIndependentCanonicalCopy(t *testing.T) {
	first := StandardStages()
	first[0] = CacheStage
	second := StandardStages()
	if second[0] != ParserStage {
		t.Fatalf("StandardStages() exposed internal order: got %q", second[0])
	}
	if !reflect.DeepEqual(second, []Stage{ParserStage, SemanticStage, QueryStage, GeneratorStage, CacheStage}) {
		t.Fatalf("StandardStages() = %#v", second)
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
