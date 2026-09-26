package generation

import "errors"

const JEVPolicyEvaluationSchemaPart01 = "gooo/generation/jev-policy-evaluation/v1"

type JEVPolicyOutcomePart01 string

const (
    JEVPolicyAcceptPart01  JEVPolicyOutcomePart01 = "ACCEPT"
    JEVPolicyDeferPart01   JEVPolicyOutcomePart01 = "DEFER"
    JEVPolicyRejectPart01  JEVPolicyOutcomePart01 = "REJECT"
    JEVPolicyUnknownPart01 JEVPolicyOutcomePart01 = "UNKNOWN"
)

type JEVDecisionPolicyPart01 struct {
    PolicyIdentity string  `json:"policy_identity"`
    AcceptDecision string  `json:"accept_decision"`
    RejectDecision string  `json:"reject_decision"`
    MinConfidence float64 `json:"min_confidence"`
}

type JEVPolicyEvaluationPart01 struct {
    Schema         string                `json:"schema"`
    ReceiptDigest  string                `json:"receipt_digest"`
    PolicyIdentity string                `json:"policy_identity"`
    Outcome        JEVPolicyOutcomePart01 `json:"outcome"`
    Reason         string                `json:"reason"`
    DecisionDigest string                `json:"decision_digest"`
}

func EvaluateJEVDecisionPolicyPart01(receipt JEVDecisionBoundaryPart01, policy JEVDecisionPolicyPart01) (JEVPolicyEvaluationPart01, error) {
    if !receipt.ValidPart01() {
        return JEVPolicyEvaluationPart01{}, errors.New("cannot evaluate invalid JEV receipt")
    }
    if policy.PolicyIdentity == "" || policy.AcceptDecision == "" || policy.RejectDecision == "" || policy.AcceptDecision == policy.RejectDecision || policy.MinConfidence < 0 || policy.MinConfidence > 1 {
        return JEVPolicyEvaluationPart01{}, errors.New("invalid JEV decision policy")
    }
    if receipt.PolicyIdentity != policy.PolicyIdentity {
        return JEVPolicyEvaluationPart01{}, errors.New("JEV receipt and policy identity differ")
    }
    outcome := JEVPolicyEvaluationPart01{
        Schema: JEVPolicyEvaluationSchemaPart01,
        ReceiptDigest: receipt.DecisionDigest,
        PolicyIdentity: policy.PolicyIdentity,
        Outcome: JEVPolicyUnknownPart01,
        Reason: "decision not eligible for deterministic policy",
    }
    switch {
    case receipt.Status == JEVDecisionUnknownPart01:
        outcome.Reason = receipt.Reason
    case receipt.Confidence < policy.MinConfidence:
        outcome.Outcome = JEVPolicyDeferPart01
        outcome.Reason = "confidence below policy threshold"
    case receipt.Decision == policy.AcceptDecision:
        outcome.Outcome = JEVPolicyAcceptPart01
        outcome.Reason = "decision matched policy accept branch"
    case receipt.Decision == policy.RejectDecision:
        outcome.Outcome = JEVPolicyRejectPart01
        outcome.Reason = "decision matched policy reject branch"
    }
    outcome.DecisionDigest = digestJEVPart01(outcome.decisionViewPart01())
    return outcome, nil
}

func (e JEVPolicyEvaluationPart01) ValidPart01() bool {
    if e.Schema != JEVPolicyEvaluationSchemaPart01 || !validDigestJEVPart01(e.ReceiptDigest) || e.PolicyIdentity == "" || e.Reason == "" || !validDigestJEVPart01(e.DecisionDigest) {
        return false
    }
    switch e.Outcome {
    case JEVPolicyAcceptPart01, JEVPolicyDeferPart01, JEVPolicyRejectPart01, JEVPolicyUnknownPart01:
    default:
        return false
    }
    return e.DecisionDigest == digestJEVPart01(e.decisionViewPart01())
}

func (e JEVPolicyEvaluationPart01) decisionViewPart01() any {
    return struct {
        Schema         string                `json:"schema"`
        ReceiptDigest  string                `json:"receipt_digest"`
        PolicyIdentity string                `json:"policy_identity"`
        Outcome        JEVPolicyOutcomePart01 `json:"outcome"`
        Reason         string                `json:"reason"`
    }{e.Schema, e.ReceiptDigest, e.PolicyIdentity, e.Outcome, e.Reason}
}