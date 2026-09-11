package selfimprovementloop

import (
	"sort"
	"strings"
)

type pairEvaluation struct {
	Decision string
	Reason   string
	Matched  bool
}

func compareExactPair(in Input) pairEvaluation {
	if len(in.Pair.Before) == 0 || len(in.Pair.After) == 0 {
		return pairEvaluation{Decision: DecisionUnknown, Reason: "no matching integer before/after pair"}
	}
	before := make(map[string]MetricSample, len(in.Pair.Before))
	after := make(map[string]MetricSample, len(in.Pair.After))
	for _, sample := range in.Pair.Before {
		if !sampleContextMatches(in, sample) {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "before sample context differs"}
		}
		key := sampleKey(sample)
		if _, exists := before[key]; exists {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "duplicate before sample"}
		}
		before[key] = sample
	}
	for _, sample := range in.Pair.After {
		if !sampleContextMatches(in, sample) {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "after sample context differs"}
		}
		key := sampleKey(sample)
		if _, exists := after[key]; exists {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "duplicate after sample"}
		}
		after[key] = sample
	}
	keys := make([]string, 0, len(before))
	for key := range before {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, exists := after[key]; exists {
			return pairEvaluation{Decision: DecisionClosed, Reason: "EXACT_INTEGER_PAIR_MATCHED", Matched: true}
		}
	}
	return pairEvaluation{Decision: DecisionUnknown, Reason: "no matching integer before/after pair"}
}

func sampleContextMatches(in Input, sample MetricSample) bool {
	return strings.TrimSpace(sample.Scenario) == in.Scenario &&
		strings.TrimSpace(sample.SourceDigest) == in.SourceDigest &&
		strings.TrimSpace(sample.ToolchainDigest) == in.ToolchainDigest &&
		strings.TrimSpace(sample.Metric) != ""
}

func sampleKey(sample MetricSample) string {
	return sample.Scenario + "\x00" + sample.SourceDigest + "\x00" + sample.ToolchainDigest + "\x00" + sample.Metric
}
