package predecessorselection

import "testing"

func TestSelectFailsClosedWithoutCanonicalCandidate(t *testing.T) {
	input := Input{
		Repository:     "owner/repository",
		CurrentHeadSHA: "2222222222222222222222222222222222222222",
		PredecessorSHA: "1111111111111111111111111111111111111111",
		Branch:         "dev",
		Workflow:       "Transformation effect ledger",
	}
	result, err := Select(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Report.Decision != DecisionFailClosed || result.Report.Reason != ReasonNotFound {
		t.Fatalf("missing predecessor did not fail closed: %+v", result.Report)
	}
}

func TestProducerConformanceIsMetricScoped(t *testing.T) {
	candidate := Candidate{RunAttempt: 1, Conclusion: "failure",
		ProducerJobID: 7, ProducerJobRunAttempt: 1,
		ProducerJobName: ProducerJobName, ProducerJobStatus: "completed",
		ProducerJobConclusion: "success", ProducerJobMatches: 1}
	if !producerConformant(candidate) {
		t.Fatal("unrelated workflow failure contaminated the metric-scoped producer")
	}
	summary := Summary{ObservedCandidates: 1, ExactHeadCandidates: 1,
		CanonicalCandidates: 1, SuccessfulCandidates: 0,
		ProducerConformantCandidates: 1, AvailableCandidates: 1, ValidCandidates: 1}
	if reason := failureReason(summary); reason != "" {
		t.Fatalf("metric-scoped producer was rejected: %s", reason)
	}
}

func TestProducerConformanceFailsClosedWhenUnknown(t *testing.T) {
	summary := Summary{ObservedCandidates: 1, ExactHeadCandidates: 1,
		CanonicalCandidates: 1, SuccessfulCandidates: 1}
	if reason := failureReason(summary); reason != ReasonProducer {
		t.Fatalf("unknown producer reason = %s, want %s", reason, ReasonProducer)
	}
}
