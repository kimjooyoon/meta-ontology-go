package feedbackpredecessor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFeedbackPredecessorUseCases(t *testing.T) {
	path := filepath.Join("..", "..", "..", "examples", "feedback-predecessor-cycle", "usecases.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture useCaseFile
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Schema != "gooo/meta-feedback-predecessor-use-cases/v1" {
		t.Fatalf("schema = %q", fixture.Schema)
	}
	for _, usecase := range fixture.Cases {
		t.Run(usecase.ID, func(t *testing.T) {
			if usecase.Actor == "" || usecase.OperatorAction == "" ||
				usecase.Trigger == "pull_request" && usecase.Cause != "base_sha" ||
				usecase.Trigger == "push" && usecase.Cause != "before_sha" {
				t.Fatalf("causal identity = %#v", usecase)
			}
			input := predecessorFixture()
			applyUseCase(&input, usecase.CandidateState)
			report, err := Select(input)
			if err != nil {
				t.Fatal(err)
			}
			values, satisfied, proofs := map[string]int{}, 0, map[string]bool{}
			for _, indicator := range report.Indicators {
				values[indicator.MetricID] = indicator.Value
				proofs[indicator.ProofChoice] = true
				if indicator.Satisfied {
					satisfied++
				}
			}
			if report.Decision != usecase.ExpectedDecision ||
				report.Reason != usecase.ExpectedReason ||
				report.Resolution != usecase.ExpectedResolution ||
				report.PromotionAuthorized != usecase.ExpectedPromotion ||
				values["gooo.metric.meta.predecessor-feedback-readiness.coverage-bps.v1"] != usecase.ExpectedReadinessBPS ||
				values["gooo.metric.meta.predecessor-cycle-continuity.coverage-bps.v1"] != usecase.ExpectedContinuityBPS ||
				values["gooo.metric.meta.predecessor-ambiguity.guardrail.v1"] != usecase.ExpectedAmbiguity ||
				values["gooo.metric.meta.predecessor-observer-writes.guardrail.v1"] != usecase.ExpectedWrites ||
				satisfied != usecase.ExpectedSatisfied || len(proofs) != 3 {
				t.Fatalf("use case %s report = %#v", usecase.ID, report)
			}
		})
	}
}
