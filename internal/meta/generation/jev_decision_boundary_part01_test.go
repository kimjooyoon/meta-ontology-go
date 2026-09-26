package generation

import (
	"strings"
	"testing"
)

func TestJEVDecisionBoundaryPart01PassAndTamper(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	receipt, err := NewJEVDecisionBoundaryPart01(JEVDecisionBoundaryInputPart01{
		StateDigest:       digest,
		QuestionDigest:    digest,
		ModelIdentity:     "jev-1.13.0",
		PolicyIdentity:    "gooo/self-improvement/v1",
		ResponseDigest:    digest,
		DecisionKind:      JEVDecisionChoicePart01,
		Decision:          "defer",
		Status:            JEVDecisionPassPart01,
		Confidence:        0.8,
		MissingStageIndex: -1,
	})
	if err != nil {
		t.Fatalf("construct receipt: %v", err)
	}
	if !receipt.ValidPart01() {
		t.Fatal("expected valid receipt")
	}
	receipt.ResponseDigest = "sha256:" + strings.Repeat("b", 64)
	if receipt.ValidPart01() {
		t.Fatal("tampered response digest must be rejected")
	}
}

func TestJEVDecisionBoundaryPart01UnknownPreservesFrontier(t *testing.T) {
	digest := "sha256:" + strings.Repeat("c", 64)
	receipt, err := NewJEVDecisionBoundaryPart01(JEVDecisionBoundaryInputPart01{
		StateDigest:       digest,
		QuestionDigest:    digest,
		ModelIdentity:     "jev-1.13.0",
		PolicyIdentity:    "gooo/self-improvement/v1",
		ResponseDigest:    digest,
		DecisionKind:      JEVDecisionNoulPart01,
		Decision:          "unknown",
		Status:            JEVDecisionUnknownPart01,
		Reason:            "action observation is absent",
		Confidence:        0.5,
		MissingStageIndex: 2,
	})
	if err != nil {
		t.Fatalf("construct unknown receipt: %v", err)
	}
	if !receipt.ValidPart01() || receipt.MissingStageIndex != 2 {
		t.Fatal("expected UNKNOWN receipt to preserve frontier")
	}
}
