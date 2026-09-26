package generation

import (
	"strings"
	"testing"
)

func TestEvaluateJEVDecisionPolicyPart01DeterministicBranches(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	receipt, err := NewJEVDecisionBoundaryPart01(JEVDecisionBoundaryInputPart01{
		StateDigest: digest, QuestionDigest: digest, ModelIdentity: "jev-1.13.0",
		PolicyIdentity: "gooo/self-improvement/v1", ResponseDigest: digest,
		DecisionKind: JEVDecisionChoicePart01, Decision: "accept",
		Status: JEVDecisionPassPart01, Confidence: 0.8, MissingStageIndex: -1,
	})
	if err != nil {
		t.Fatalf("construct receipt: %v", err)
	}
	policy := JEVDecisionPolicyPart01{PolicyIdentity: "gooo/self-improvement/v1", AcceptDecision: "accept", RejectDecision: "reject", MinConfidence: 0.7}
	evaluation, err := EvaluateJEVDecisionPolicyPart01(receipt, policy)
	if err != nil || !evaluation.ValidPart01() || evaluation.Outcome != JEVPolicyAcceptPart01 {
		t.Fatalf("expected ACCEPT evaluation: %#v, %v", evaluation, err)
	}
	receipt.Confidence = 0.2
	receipt.EvidencePrefixDigest = digestJEVPart01(receipt.evidencePrefixPart01())
	receipt.DecisionDigest = digestJEVPart01(receipt.decisionViewPart01())
	evaluation, err = EvaluateJEVDecisionPolicyPart01(receipt, policy)
	if err != nil || evaluation.Outcome != JEVPolicyDeferPart01 {
		t.Fatalf("expected DEFER evaluation: %#v, %v", evaluation, err)
	}
}

func TestEvaluateJEVDecisionPolicyPart01NeverPromotesUnknown(t *testing.T) {
	digest := "sha256:" + strings.Repeat("b", 64)
	receipt, err := NewJEVDecisionBoundaryPart01(JEVDecisionBoundaryInputPart01{
		StateDigest: digest, QuestionDigest: digest, ModelIdentity: "jev-1.13.0",
		PolicyIdentity: "gooo/self-improvement/v1", ResponseDigest: digest,
		DecisionKind: JEVDecisionChoicePart01, Decision: "accept",
		Status: JEVDecisionUnknownPart01, Reason: "reverse observation absent", Confidence: 1, MissingStageIndex: 3,
	})
	if err != nil {
		t.Fatalf("construct UNKNOWN receipt: %v", err)
	}
	evaluation, err := EvaluateJEVDecisionPolicyPart01(receipt, JEVDecisionPolicyPart01{PolicyIdentity: "gooo/self-improvement/v1", AcceptDecision: "accept", RejectDecision: "reject", MinConfidence: 0.1})
	if err != nil || !evaluation.ValidPart01() || evaluation.Outcome != JEVPolicyUnknownPart01 {
		t.Fatalf("UNKNOWN must remain UNKNOWN: %#v, %v", evaluation, err)
	}
	evaluation.Outcome = JEVPolicyAcceptPart01
	if evaluation.ValidPart01() {
		t.Fatal("tampered UNKNOWN evaluation must be rejected")
	}
}
