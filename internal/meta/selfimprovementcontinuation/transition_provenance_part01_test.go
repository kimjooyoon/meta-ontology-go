package selfimprovementcontinuation

import (
	"strings"
	"testing"
)

func TestObserveContinuationTransitionPreservesImprovementAndRegression(t *testing.T) {
	beforeDigest := "sha256:" + strings.Repeat("a", 64)
	afterDigest := "sha256:" + strings.Repeat("b", 64)
	cases := []struct {
		name      string
		before    Decision
		after     Decision
		direction string
		reason    string
	}{
		{name: "improved", before: DecisionUnknown, after: DecisionClosed, direction: ContinuationTransitionImproved, reason: "CONTINUATION_RESOLUTION_CLOSED"},
		{name: "regressed", before: DecisionClosed, after: DecisionRefuted, direction: ContinuationTransitionRegressed, reason: "CONTINUATION_RESOLUTION_OPENED"},
		{name: "stable", before: DecisionUnknown, after: DecisionUnknown, direction: ContinuationTransitionStable, reason: "CONTINUATION_RESOLUTION_UNCHANGED"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			observation := ObserveContinuationTransition(
				ContinuationResolution{Digest: beforeDigest, Decision: testCase.before, Metrics: Metrics{ClosedCases: 1, UnknownCases: 2, RefutedCases: 3}},
				ContinuationResolution{Digest: afterDigest, Decision: testCase.after, Metrics: Metrics{ClosedCases: 2, UnknownCases: 3, RefutedCases: 4}},
			)
			if observation.Direction != testCase.direction || observation.Reason != testCase.reason {
				t.Fatalf("observation = %#v", observation)
			}
			if err := observation.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestObserveContinuationTransitionRejectsSnapshotTampering(t *testing.T) {
	beforeDigest := "sha256:" + strings.Repeat("a", 64)
	afterDigest := "sha256:" + strings.Repeat("b", 64)
	observation := ObserveContinuationTransition(
		ContinuationResolution{Digest: beforeDigest, Decision: DecisionClosed, Metrics: Metrics{ClosedCases: 2}},
		ContinuationResolution{Digest: afterDigest, Decision: DecisionUnknown, Metrics: Metrics{ClosedCases: 1}},
	)
	observation.AfterClosedCases = 9
	if err := observation.Validate(); err == nil {
		t.Fatal("tampered continuation snapshot was accepted")
	}
}

func TestObserveContinuationTransitionPreservesUnknownEvidence(t *testing.T) {
	observation := ObserveContinuationTransition(
		ContinuationResolution{Decision: DecisionClosed},
		ContinuationResolution{Decision: DecisionUnknown},
	)
	if observation.Direction != ContinuationTransitionUnknown ||
		observation.Reason != "CONTINUATION_TRANSITION_UNKNOWN" {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("unknown Validate() error = %v", err)
	}
}
