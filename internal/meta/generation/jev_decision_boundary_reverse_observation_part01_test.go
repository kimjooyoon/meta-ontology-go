package generation

import (
    "strings"
    "testing"
)

func TestAttachJEVActionObservationPart01RecomputesReceipt(t *testing.T) {
    digest := "sha256:" + strings.Repeat("a", 64)
    receipt, err := NewJEVDecisionBoundaryPart01(JEVDecisionBoundaryInputPart01{
        StateDigest:        digest,
        QuestionDigest:     digest,
        ModelIdentity:      "jev-1.13.0",
        PolicyIdentity:     "gooo/self-improvement/v1",
        ResponseDigest:     digest,
        DecisionKind:       JEVDecisionChoicePart01,
        Decision:           "defer",
        Status:             JEVDecisionPassPart01,
        Confidence:         0.8,
        MissingStageIndex:  -1,
    })
    if err != nil {
        t.Fatalf("construct receipt: %v", err)
    }
    actionDigest := "sha256:" + strings.Repeat("b", 64)
    observed, err := AttachJEVActionObservationPart01(receipt, actionDigest)
    if err != nil {
        t.Fatalf("attach action observation: %v", err)
    }
    if !observed.ValidPart01() || observed.ActionObservationDigest != actionDigest {
        t.Fatal("expected re-bound receipt to verify")
    }
    observed.ActionObservationDigest = "sha256:" + strings.Repeat("c", 64)
    if observed.ValidPart01() {
        t.Fatal("tampered action observation must be rejected")
    }
}