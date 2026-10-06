package bodycodegen

import (
	"slices"
	"testing"
)

func TestAutomaticBodyFillProbesAreDeterministicBoundedAndTrainingDerived(t *testing.T) {
	training := make([]IRBodyFillTestCase, 1001)
	for index := range training {
		training[index].Input = int64(index * 2)
	}
	all := deriveIRBodyFillBehavioralProbeInputs(training)
	first := retainIRBodyFillBehavioralProbes(all, irBodyFillBehavioralProbeCap)
	second := retainIRBodyFillBehavioralProbes(deriveIRBodyFillBehavioralProbeInputs(training), irBodyFillBehavioralProbeCap)
	if len(all) <= irBodyFillBehavioralProbeCap || len(first) != irBodyFillBehavioralProbeCap || !slices.Equal(first, second) {
		t.Fatalf("automatic probe set was not deterministically capped: full=%d first=%d second=%d", len(all), len(first), len(second))
	}
	if first[0] != all[0] || first[len(first)-1] != all[len(all)-1] {
		t.Fatalf("bounded probe sampling omitted the domain edges: first=%d..%d full=%d..%d", first[0], first[len(first)-1], all[0], all[len(all)-1])
	}
	trainingInputs := make(map[int64]struct{}, len(training))
	for _, testCase := range training {
		trainingInputs[testCase.Input] = struct{}{}
	}
	for index, input := range all {
		if _, isTraining := trainingInputs[input]; isTraining || input%2 == 0 {
			t.Fatalf("automatic probe %d reused a training input instead of deriving an unseen boundary: %d", index, input)
		}
		if index > 0 && all[index-1] >= input {
			t.Fatalf("automatic probes are not unique and sorted: %d then %d", all[index-1], input)
		}
	}
}

func TestAutomaticBodyFillProbesDoNotOverflowAtInt64Edges(t *testing.T) {
	minInt64 := int64(-1 << 63)
	maxInt64 := int64(1<<63 - 1)
	probes := deriveIRBodyFillBehavioralProbeInputs([]IRBodyFillTestCase{{Input: minInt64}, {Input: maxInt64}})
	for index, input := range probes {
		if input == minInt64 || input == maxInt64 {
			t.Fatalf("automatic probes included a training endpoint: %d", input)
		}
		if index > 0 && probes[index-1] >= input {
			t.Fatalf("edge probes are not unique and sorted: %d then %d", probes[index-1], input)
		}
	}
	for _, expected := range []int64{minInt64 + 1, -2, -1, 0, 1, maxInt64 - 1} {
		if !slices.Contains(probes, expected) {
			t.Errorf("edge probe set is missing %d: %v", expected, probes)
		}
	}
}
