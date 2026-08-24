package externalconformanceactivation

import (
	"encoding/json"
	"reflect"
	"testing"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestExactActivation(t *testing.T) {
	first, replay := Evaluate(EmbeddedInput(candidateSHA)), Evaluate(EmbeddedInput(candidateSHA))
	if !reflect.DeepEqual(first, replay) {
		t.Fatal("activation replay differs")
	}
	if first.Decision != DecisionApplied || first.Resolution != ResolutionExact ||
		first.EnforcementEffect != EffectApply || first.TransitionApplied != 1 || first.RepositoryWrites != 0 {
		t.Fatalf("receipt=%+v", first)
	}
	s := first.Summary
	if s.BeforeOperating != 11 || s.AfterOperating != 12 || s.BeforeCoverageBPS != 9166 ||
		s.AfterCoverageBPS != 10000 || s.CapsulesExact != 3 || s.PredecessorSemanticsBPS != 10000 ||
		s.EligibilityIndicatorsSatisfied != 18 || s.ParentCompleted != 6 || s.ParentTotal != 8 ||
		s.ParentKnownFailures != 2 || s.SelectedCompleted != 10 || s.SelectedTotal != 10 || s.ExternalExecutions != 4 {
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
		{"unavailable", ResolutionUnknown, ReasonUnavailable, func(input *Input) { input.Assurance = nil }},
		{"digest-mismatch", ResolutionInvariant, ReasonDigestMismatch, func(input *Input) { input.Merge = append(input.Merge, '\n') }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := EmbeddedInput(candidateSHA)
			test.mutate(&input)
			assertBlocked(t, Evaluate(input), test.resolution, test.reason)
		})
	}
}

func TestUnknownTopValuesNeverActivate(t *testing.T) {
	tests := []struct{ name, capsule, field, value, reason string }{
		{"eligibility-unknown", "eligibility", "decision", "UNKNOWN", ReasonEligibilityUnknown},
		{"eligibility-fixed-point", "eligibility", "decision", "FIXED_POINT", ReasonEligibilityUnknown},
		{"eligibility-unrecognized", "eligibility", "decision", "SURPRISE", ReasonEligibilityUnknown},
		{"assurance-unknown", "assurance", "candidate_decision", "UNKNOWN", ReasonAssuranceUnknown},
		{"merge-unknown", "merge", "state", "UNKNOWN", ReasonMergeUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := EmbeddedInput(candidateSHA)
			raw := map[string]any{}
			target := capsule(&input, test.capsule)
			if json.Unmarshal(*target, &raw) != nil {
				t.Fatal("decode capsule")
			}
			raw[test.field] = test.value
			*target, _ = json.Marshal(raw)
			assertBlocked(t, Evaluate(input), ResolutionUnknown, test.reason)
		})
	}
}
