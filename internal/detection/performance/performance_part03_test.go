package performance

import (
	"reflect"
	"strings"
	"testing"
)

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
