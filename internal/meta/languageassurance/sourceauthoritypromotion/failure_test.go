package sourceauthoritypromotion

import (
	"encoding/json"
	"testing"
)

func TestEvaluateFailsClosed(t *testing.T) {
	tests := []struct {
		name, reason string
		mutate       func(*Input)
	}{
		{name: "upstream-not-exact", reason: ReasonUpstreamNotExact, mutate: mutateUpstreamUnknown},
		{name: "baseline-already-operating", reason: ReasonBaselineState, mutate: mutateBaselineOperating},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validInput(t)
			test.mutate(&input)
			report := Evaluate(input)
			if report.Decision != DecisionBlock || report.Resolution != ResolutionInvariantOnly || report.Reason != test.reason {
				t.Fatalf("unknown evidence was not blocked: %#v", report)
			}
			if report.PromotionApplied != 0 || report.RepositoryWrites != 0 { t.Fatalf("blocked report had effects: %#v", report) }
		})
	}
}

func mutateUpstreamUnknown(input *Input) {
	var document upstreamDocument
	_ = json.Unmarshal(input.UpstreamJSON, &document)
	document.Decision = "UNKNOWN"
	input.UpstreamJSON, _ = json.Marshal(document)
}

func mutateBaselineOperating(input *Input) {
	var document assuranceDocument
	_ = json.Unmarshal(input.AssuranceJSON, &document)
	for index := range document.Obligations {
		if document.Obligations[index].MetricID == SourceMetric {
			document.Obligations[index].Status, document.Obligations[index].Resolution = "OPERATING", ResolutionExact
		}
	}
	input.AssuranceJSON, _ = json.Marshal(document)
}
