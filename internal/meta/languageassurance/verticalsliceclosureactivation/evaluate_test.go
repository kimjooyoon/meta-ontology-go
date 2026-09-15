package verticalsliceclosureactivation

import (
	"encoding/json"
	"reflect"
	"testing"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestExactActivation(t *testing.T) {
	first, second := Evaluate(EmbeddedInput(candidateSHA)), Evaluate(EmbeddedInput(candidateSHA))
	if !reflect.DeepEqual(first, second) {
		t.Fatal("activation replay differs")
	}
	if first.Decision != DecisionApplied || first.Resolution != ResolutionExact || first.EnforcementEffect != EffectApply || first.TransitionApplied != 1 || first.RepositoryWrites != 0 {
		t.Fatalf("receipt=%+v", first)
	}
	s := first.Summary
	if s.BeforeOperating != 10 || s.AfterOperating != 11 || s.BeforeCoverageBPS != 8333 || s.AfterCoverageBPS != 9166 ||
		s.CapsulesExact != 2 || s.PredecessorSemanticsBPS != 10000 || s.BoundariesSatisfied != 6 || s.LinksSatisfied != 12 ||
		s.EligibilityIndicatorsSatisfied != 8 || s.MetaOperationsObserved != 6 {
		t.Fatalf("summary=%+v", s)
	}
	metricID, operation, err := OperatingOperation()
	if err != nil || metricID != MetricID || operation != MetaOperation {
		t.Fatalf("operation=%s/%s err=%v", metricID, operation, err)
	}
}

func TestActivationFailsClosed(t *testing.T) {
	tests := []struct {
		name, resolution, reason string
		mutate                   func(*Input)
	}{
		{name: "unavailable", resolution: ResolutionUnknown, reason: ReasonUnavailable, mutate: func(input *Input) { input.Assurance = nil }},
		{name: "digest-mismatch", resolution: ResolutionInvariant, reason: ReasonDigestMismatch, mutate: func(input *Input) { input.Assurance = append(input.Assurance, '\n') }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := EmbeddedInput(candidateSHA)
			test.mutate(&input)
			receipt := Evaluate(input)
			if receipt.Decision != DecisionFailClosed || receipt.Resolution != test.resolution || receipt.EnforcementEffect != EffectBlock || receipt.Reason != test.reason || receipt.TransitionApplied != 0 || receipt.RepositoryWrites != 0 {
				t.Fatalf("receipt=%+v", receipt)
			}
		})
	}
}

func TestUnknownEligibilityLowersResolution(t *testing.T) {
	input := EmbeddedInput(candidateSHA)
	var report map[string]any
	if err := json.Unmarshal(input.Eligibility, &report); err != nil {
		t.Fatal(err)
	}
	report["decision"], report["resolution"] = "UNKNOWN", "UNKNOWN"
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	input.Eligibility = raw
	receipt := Evaluate(input)
	if receipt.Decision != DecisionFailClosed || receipt.Resolution != ResolutionUnknown || receipt.EnforcementEffect != EffectBlock || receipt.Reason != ReasonEligibilityUnknown || receipt.TransitionApplied != 0 || receipt.RepositoryWrites != 0 {
		t.Fatalf("receipt=%+v", receipt)
	}
}
