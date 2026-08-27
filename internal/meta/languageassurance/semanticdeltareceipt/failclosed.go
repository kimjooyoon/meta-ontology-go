package semanticdeltareceipt

import "encoding/hex"

func unknownReceipt(receipt Receipt, before, after projectedSource, beforeErr, afterErr error, coordinates ...string) Receipt {
	stage, step, reason := UnavailableStage, UnavailableStep, ReasonUnavailable
	if len(coordinates) == 3 {
		stage, step, reason = coordinates[0], coordinates[1], coordinates[2]
	}
	receipt.StructuralDelta = StructuralDelta{Status: "UNKNOWN"}
	receipt.SemanticComponentDelta = SemanticComponentDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN"}
	receipt.ModeledSemanticComponents, receipt.TotalSemanticComponents, receipt.SemanticCoverageBPS = 0, TotalComponentCount, 0
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, SemanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFailClosed, ResolutionLower, ClassIndeterminate, reason
	receipt.Stage, receipt.Step = stage, step
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, stage, step, reason)
	receipt.TransitionCount = len(receipt.ClaimTransitions)
	if receipt.TransitionCount > 0 {
		receipt.TransitionHeadDigest = receipt.ClaimTransitions[receipt.TransitionCount-1].TransitionDigest
	}
	if beforeErr == nil && afterErr != nil {
		receipt.After.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	if beforeErr != nil && afterErr == nil {
		receipt.Before.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	sealReceipt(&receipt)
	return receipt
}

func subjectUnknown(receipt Receipt, before, after projectedSource) Receipt {
	receipt.StructuralDelta = StructuralDelta{Status: "UNKNOWN"}
	receipt.SemanticComponentDelta = SemanticComponentDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN"}
	receipt.ModeledSemanticComponents, receipt.TotalSemanticComponents, receipt.SemanticCoverageBPS = 0, TotalComponentCount, 0
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, SemanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFailClosed, ResolutionLower, ClassIndeterminate, ReasonSubject
	receipt.Stage, receipt.Step = SubjectStage, SubjectStep
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, SubjectStage, SubjectStep, ReasonSubject)
	receipt.TransitionCount = len(receipt.ClaimTransitions)
	if receipt.TransitionCount > 0 {
		receipt.TransitionHeadDigest = receipt.ClaimTransitions[receipt.TransitionCount-1].TransitionDigest
	}
	sealReceipt(&receipt)
	return receipt
}

func ambiguousReceipt(receipt Receipt, before, after projectedSource) Receipt {
	receipt.StructuralDelta = structuralDelta(before, after)
	receipt.SemanticComponentDelta = semanticComponentDelta(before, after)
	receipt.SemanticClaimDelta = claimDelta(before, after)
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, SemanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFailClosed, ResolutionLower, ClassIndeterminate, ReasonAmbiguous
	receipt.Stage, receipt.Step = "claim-delta", "match-claims"
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, receipt.Stage, receipt.Step, receipt.Reason)
	receipt.TransitionCount = len(receipt.ClaimTransitions)
	if receipt.TransitionCount > 0 {
		receipt.TransitionHeadDigest = receipt.ClaimTransitions[receipt.TransitionCount-1].TransitionDigest
	}
	sealReceipt(&receipt)
	return receipt
}

func unmodeledReceipt(receipt Receipt, before, after projectedSource) Receipt {
	receipt.StructuralDelta = structuralDelta(before, after)
	receipt.SemanticComponentDelta = SemanticComponentDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN", Reason: ReasonUnmodeled}
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, SemanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFailClosed, ResolutionLower, ClassIndeterminate, ReasonUnmodeled
	receipt.Stage, receipt.Step = "semantic-projection", "compare-stable-hash"
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, receipt.Stage, receipt.Step, receipt.Reason)
	receipt.TransitionCount = len(receipt.ClaimTransitions)
	if receipt.TransitionCount > 0 {
		receipt.TransitionHeadDigest = receipt.ClaimTransitions[receipt.TransitionCount-1].TransitionDigest
	}
	sealReceipt(&receipt)
	return receipt
}

func validSubject(value string) bool {
	if len(value) != 40 || value != stringLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func stringLower(value string) string {
	result := []byte(value)
	for index, char := range result {
		if char >= 'A' && char <= 'F' {
			result[index] += 'a' - 'A'
		}
	}
	return string(result)
}
