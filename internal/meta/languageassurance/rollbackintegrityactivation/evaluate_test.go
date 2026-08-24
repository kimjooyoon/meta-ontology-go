package rollbackintegrityactivation

import (
	"encoding/json"
	"reflect"
	"testing"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestExactActivation(t *testing.T) {
	first := Evaluate(EmbeddedInput(candidateSHA))
	second := Evaluate(EmbeddedInput(candidateSHA))
	if !reflect.DeepEqual(first, second) {
		t.Fatal("activation replay differs")
	}
	if first.Decision != DecisionApplied || first.Resolution != ResolutionExact || first.EnforcementEffect != EffectApply ||
		first.TransitionApplied != 1 || first.RepositoryWrites != 0 {
		t.Fatalf("receipt=%+v", first)
	}
	s := first.Summary
	if s.BeforeOperating != 9 || s.AfterOperating != 10 || s.BeforeCoverageBPS != 7500 || s.AfterCoverageBPS != 8333 ||
		s.CapsulesExact != 2 || s.PredecessorSemanticsBPS != 10000 || s.ShadowCasesPassed != 7 ||
		s.ShadowReplaysExact != 2 || s.MetaOperationsObserved != 6 || s.MetaOperationCoverageBPS != 10000 {
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
			if receipt.Decision != DecisionFailClosed || receipt.Resolution != test.resolution || receipt.EnforcementEffect != EffectBlock ||
				receipt.Reason != test.reason || receipt.TransitionApplied != 0 || receipt.RepositoryWrites != 0 {
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
	report["decision"] = "UNKNOWN"
	report["resolution"] = "UNKNOWN"
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	input.Eligibility = raw
	receipt := Evaluate(input)
	if receipt.Decision != DecisionFailClosed || receipt.Resolution != ResolutionUnknown ||
		receipt.EnforcementEffect != EffectBlock || receipt.Reason != ReasonEligibilityUnknown ||
		receipt.TransitionApplied != 0 || receipt.RepositoryWrites != 0 {
		t.Fatalf("receipt=%+v", receipt)
	}
}

func TestActivationDenominatorAndMetaBindings(t *testing.T) {
	if cases := Denominator(); len(cases) != 4 {
		t.Fatalf("cases=%d", len(cases))
	}
	receipt := Evaluate(EmbeddedInput(candidateSHA))
	classes, proofs := map[string]int{}, map[string]int{}
	for _, indicator := range receipt.Indicators {
		classes[indicator.Class]++
		proofs[indicator.ProofChoice]++
	}
	if len(receipt.Indicators) != 6 || classes["OUTCOME"] != 1 || classes["DRIVER"] != 2 || classes["GUARDRAIL"] != 3 ||
		proofs["FOUNDATION"] != 3 || proofs["COHERENCE"] != 2 || proofs["REGRESSION"] != 1 ||
		len(receipt.MetaOperations) != 6 {
		t.Fatalf("indicators=%+v meta=%+v", receipt.Indicators, receipt.MetaOperations)
	}
}
