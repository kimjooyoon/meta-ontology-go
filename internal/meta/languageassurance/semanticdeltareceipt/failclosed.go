package semanticdeltareceipt

import "encoding/hex"

func unknownReceipt(receipt Receipt, before, after projectedSource, beforeErr, afterErr error) Receipt {
	receipt.StructuralDelta = StructuralDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN"}
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, SemanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFailClosed, ResolutionUnknown, ClassIndeterminate, ReasonUnavailable
	receipt.Stage, receipt.Step = "produce", "fail-closed"
	receipt.ClaimTransitions = transitions(before, after, ClassIndeterminate, ReasonUnavailable)
	if beforeErr == nil && afterErr != nil {
		receipt.After.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	if beforeErr != nil && afterErr == nil {
		receipt.Before.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	sealReceipt(&receipt)
	return receipt
}

func subjectUnknown(receipt Receipt) Receipt {
	receipt.StructuralDelta = StructuralDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN"}
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, SemanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFailClosed, ResolutionLower, ClassIndeterminate, ReasonSubject
	receipt.Stage, receipt.Step = SubjectStage, SubjectStep
	receipt.ClaimTransitions = []ClaimTransition{{ClaimID: boundedEquivalenceClaim, FromStatus: StatusOpen, ToStatus: StatusOpen, Stage: SubjectStage, Step: SubjectStep, Reason: ReasonSubject}}
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
