package candidateleakageactivation

import (
	"encoding/json"
	"reflect"
	"testing"

	eligibility "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/candidateleakageeligibility"
)

const candidateSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestExactActivation(t *testing.T) {
	first := Evaluate(EmbeddedInput(candidateSHA))
	second := Evaluate(EmbeddedInput(candidateSHA))
	if !reflect.DeepEqual(first, second) {
		t.Fatal("activation replay differs")
	}
	if first.Decision != DecisionApplied || first.Resolution != ResolutionExact || first.TransitionApplied != 1 || first.RepositoryWrites != 0 {
		t.Fatalf("receipt=%+v", first)
	}
	if first.Summary.BeforeOperating != 7 || first.Summary.AfterOperating != 8 || first.Summary.BeforeCoverageBPS != 5833 || first.Summary.AfterCoverageBPS != 6666 || first.Summary.CapsulesExact != 2 || first.Summary.PredecessorSemanticsBPS != 10000 {
		t.Fatalf("summary=%+v", first.Summary)
	}
	metricID, operation, err := OperatingOperation()
	if err != nil || metricID != MetricID || operation != MetaOperation {
		t.Fatalf("operation=%s/%s err=%v", metricID, operation, err)
	}
}

func TestActivationFailsClosed(t *testing.T) {
	tests := []struct {
		name, decision, resolution, reason string
		mutate                             func(*Input)
	}{
		{name: "unavailable", decision: DecisionFailClosed, resolution: ResolutionUnknown, reason: ReasonUnavailable, mutate: func(input *Input) { input.Assurance = nil }},
		{name: "digest-mismatch", decision: DecisionFailClosed, resolution: ResolutionInvariant, reason: ReasonDigestMismatch, mutate: func(input *Input) { input.Assurance = append(input.Assurance, '\n') }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := EmbeddedInput(candidateSHA)
			test.mutate(&input)
			receipt := Evaluate(input)
			if receipt.Decision != test.decision || receipt.Resolution != test.resolution || receipt.Reason != test.reason || receipt.TransitionApplied != 0 || receipt.RepositoryWrites != 0 {
				t.Fatalf("receipt=%+v", receipt)
			}
		})
	}
}

func TestUnknownEligibilityNeverActivates(t *testing.T) {
	var report eligibility.Report
	if err := json.Unmarshal(eligibilityCapsule, &report); err != nil {
		t.Fatal(err)
	}
	report.Decision = "UNKNOWN"
	if validEligibility(report) {
		t.Fatal("unknown eligibility activated")
	}
}

func TestActivationDenominator(t *testing.T) {
	if cases := Denominator(); len(cases) != 3 {
		t.Fatalf("cases=%d", len(cases))
	}
}
