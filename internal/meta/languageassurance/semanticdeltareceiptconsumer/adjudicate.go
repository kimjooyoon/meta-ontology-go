package semanticdeltareceiptconsumer

import (
	"os"
	"reflect"
)

// AdjudicateFiles reconstructs the complete expected receipt from the raw
// pair, raw meta source, checkout observation, and effects snapshots. The
// producer receipt is only compared with that reconstruction.
func AdjudicateFiles(input Input, receipt Receipt) Verdict {
	beforeRaw, beforeErr := os.ReadFile(input.BeforePath)
	afterRaw, afterErr := os.ReadFile(input.AfterPath)
	if beforeErr != nil || afterErr != nil {
		return mismatchVerdict()
	}
	expected := reconstructReceipt(input, beforeRaw, afterRaw)
	if !reflect.DeepEqual(receipt, expected) {
		return mismatchVerdict()
	}
	return Verdict{Decision: expected.Decision, Resolution: expected.Resolution, Classification: expected.Classification, Reason: expected.Reason, Passed: true, Producer: producerName, Consumer: consumerName, Stage: expected.Stage, Step: expected.Step}
}

func reconstructReceipt(input Input, beforeRaw, afterRaw []byte) Receipt {
	beforeSource, beforeErr := projectSourceSide(input.BeforePath, beforeRaw, true)
	afterSource, afterErr := projectSourceSide(input.AfterPath, afterRaw, false)
	if beforeSource.path == "" {
		beforeSource = sourceEnvelope(input.BeforePath, beforeRaw)
	}
	if afterSource.path == "" {
		afterSource = sourceEnvelope(input.AfterPath, afterRaw)
	}
	meta, metaErr := readMetaContract()
	receipt := Receipt{Schema: receiptSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA, ExpectedSubjectSHA: input.SubjectSHA, ObservedCheckoutSHA: input.ObservedCheckoutSHA, SubjectBinding: subjectBinding(input.SubjectSHA, input.ObservedCheckoutSHA), Producer: producerName, Consumer: consumerName, MetaOperation: metaOperation, ProofChoice: "FOUNDATION", Stage: "produce", Step: "separate-delta-layers", MetaSourcePath: metaSourcePath, MetaContractDigest: meta.Digest, DenominatorVersion: meta.Version, DenominatorCases: meta.DenominatorCases, ModeledSemanticComponents: modeledComponentCount, TotalSemanticComponents: totalComponentCount, DeclaredProjectionComponentKindCoverageBPS: coverageBPS(modeledComponentCount, totalComponentCount), SemanticEquivalenceClaim: semanticEquivalenceNotClaimed, Before: snapshot(beforeRaw, beforeSource, beforeErr), After: snapshot(afterRaw, afterSource, afterErr), TextualDelta: textualDelta(beforeRaw, afterRaw), Effects: observeEffects(input)}
	if metaErr != nil {
		return unknownReceipt(receipt, beforeSource, afterSource, nil, nil, "meta-source", "parse-lower", reasonMeta)
	}
	if beforeErr != nil || afterErr != nil {
		return unknownReceipt(receipt, beforeSource, afterSource, beforeErr, afterErr, unavailableStage, unavailableStep, reasonUnavailable)
	}
	if receipt.SubjectBinding != "EXACT" {
		return subjectUnknown(receipt, beforeSource, afterSource, receipt.SubjectBinding)
	}
	receipt.StructuralDelta = structuralDelta(beforeSource, afterSource)
	receipt.SemanticComponentDelta = semanticComponentDelta(beforeSource, afterSource)
	receipt.SemanticClaimDelta = claimDelta(beforeSource, afterSource)
	if receipt.SemanticClaimDelta.Status == "UNKNOWN" {
		return ambiguousReceipt(receipt, beforeSource, afterSource)
	}
	receipt.RawDecision = receipt.TextualDelta.Decision
	receipt.SemanticDecision = semanticPreserved
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = decisionFixedPoint, resolutionExact, classPreserved, reasonTextualOnly
	if beforeSource.semanticDigest != afterSource.semanticDigest && len(receipt.SemanticComponentDelta.Added)+len(receipt.SemanticComponentDelta.Removed)+len(receipt.SemanticComponentDelta.Changed) == 0 {
		return unmodeledReceipt(receipt, beforeSource, afterSource)
	}
	if hasSemanticDelta(receipt.StructuralDelta, receipt.SemanticComponentDelta, receipt.SemanticClaimDelta) {
		receipt.SemanticDecision = semanticChanged
		receipt.Decision, receipt.Classification = decisionDelta, classChanged
		if len(receipt.SemanticComponentDelta.Added)+len(receipt.SemanticComponentDelta.Removed)+len(receipt.SemanticComponentDelta.Changed) > 0 && len(receipt.StructuralDelta.AddedNodes)+len(receipt.StructuralDelta.RemovedNodes)+len(receipt.StructuralDelta.AddedFacts)+len(receipt.StructuralDelta.RemovedFacts)+len(receipt.SemanticClaimDelta.Added)+len(receipt.SemanticClaimDelta.Removed)+len(receipt.SemanticClaimDelta.Changed) == 0 {
			receipt.Reason = reasonComponentDelta
		} else {
			receipt.Reason = reasonMeaning
		}
	}
	receipt.Stage, receipt.Step = "produce", "classify"
	receipt.ClaimLedger, receipt.ClaimTransitions = claimLedger(beforeSource, afterSource, receipt.Classification, receipt.Reason)
	finishReceiptClaims(&receipt)
	sealReceipt(&receipt)
	return receipt
}

func subjectBinding(expected, observed string) string {
	if !validSubject(expected) || !validSubject(observed) {
		return "UNKNOWN"
	}
	if expected != observed {
		return "REFUTED"
	}
	return "EXACT"
}

func unknownReceipt(receipt Receipt, before, after projectedSource, beforeErr, afterErr error, coordinates ...string) Receipt {
	stage, step, reason := unavailableStage, unavailableStep, reasonUnavailable
	if len(coordinates) == 3 {
		stage, step, reason = coordinates[0], coordinates[1], coordinates[2]
	}
	receipt.StructuralDelta = StructuralDelta{Status: "UNKNOWN"}
	receipt.SemanticComponentDelta = SemanticComponentDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN"}
	receipt.ModeledSemanticComponents, receipt.TotalSemanticComponents, receipt.DeclaredProjectionComponentKindCoverageBPS = 0, totalComponentCount, 0
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, semanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reason
	receipt.Stage, receipt.Step = stage, step
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, stage, step, reason)
	finishReceiptClaims(&receipt)
	if beforeErr == nil && afterErr != nil {
		receipt.After.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	if beforeErr != nil && afterErr == nil {
		receipt.Before.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	sealReceipt(&receipt)
	return receipt
}

func subjectUnknown(receipt Receipt, before, after projectedSource, binding string) Receipt {
	receipt.StructuralDelta = StructuralDelta{Status: "UNKNOWN"}
	receipt.SemanticComponentDelta = SemanticComponentDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN"}
	receipt.ModeledSemanticComponents, receipt.TotalSemanticComponents, receipt.DeclaredProjectionComponentKindCoverageBPS = 0, totalComponentCount, 0
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, semanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification = decisionFailClosed, resolutionLower, classIndeterminate
	step, reason := subjectCoordinates(receipt.ExpectedSubjectSHA, receipt.ObservedCheckoutSHA)
	receipt.Stage, receipt.Step, receipt.SubjectBinding, receipt.Reason = "bind-subject", step, binding, reason
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, receipt.Stage, receipt.Step, receipt.Reason)
	finishReceiptClaims(&receipt)
	sealReceipt(&receipt)
	return receipt
}

func subjectCoordinates(expected, observed string) (string, string) {
	if !validSubject(expected) || !validSubject(observed) {
		if !validSubject(expected) || (observed != "" && observed != "UNKNOWN") {
			return "validate-sha", reasonSubjectSHAInvalid
		}
		return "observe-checkout-sha", reasonSubjectSHAUnavailable
	}
	if expected != observed {
		return "compare-sha", reasonSubjectSHAMismatch
	}
	return "resolve-subject", ""
}

func ambiguousReceipt(receipt Receipt, before, after projectedSource) Receipt {
	receipt.StructuralDelta = structuralDelta(before, after)
	receipt.SemanticComponentDelta = semanticComponentDelta(before, after)
	receipt.SemanticClaimDelta = claimDelta(before, after)
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, semanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonAmbiguous
	receipt.Stage, receipt.Step = "claim-delta", "match-claims"
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, receipt.Stage, receipt.Step, receipt.Reason)
	finishReceiptClaims(&receipt)
	sealReceipt(&receipt)
	return receipt
}

func unmodeledReceipt(receipt Receipt, before, after projectedSource) Receipt {
	receipt.StructuralDelta = structuralDelta(before, after)
	receipt.SemanticComponentDelta = SemanticComponentDelta{Status: "UNKNOWN"}
	receipt.SemanticClaimDelta = ClaimDelta{Status: "UNKNOWN", Reason: reasonUnmodeled}
	receipt.ModeledSemanticComponents, receipt.TotalSemanticComponents, receipt.DeclaredProjectionComponentKindCoverageBPS = 0, totalComponentCount, 0
	receipt.RawDecision, receipt.SemanticDecision = receipt.TextualDelta.Decision, semanticUnknown
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonUnmodeled
	receipt.Stage, receipt.Step = "semantic-projection", "compare-stable-hash"
	receipt.ClaimLedger, receipt.ClaimTransitions = unknownLedger(before, after, receipt.Stage, receipt.Step, receipt.Reason)
	finishReceiptClaims(&receipt)
	sealReceipt(&receipt)
	return receipt
}

func finishReceiptClaims(receipt *Receipt) {
	receipt.TransitionCount = len(receipt.ClaimTransitions)
	if receipt.TransitionCount > 0 {
		receipt.TransitionHeadDigest = receipt.ClaimTransitions[receipt.TransitionCount-1].TransitionDigest
	}
	receipt.ClaimsWithExplainedStatus, receipt.TotalClaims, receipt.ClaimStatusCoverageBPS = claimStatusCoverage(receipt.ClaimLedger, receipt.ClaimTransitions)
}

func mismatchVerdict() Verdict {
	return Verdict{Decision: decisionFailClosed, Resolution: resolutionInvariant, Classification: classIndeterminate, Reason: reasonReceipt, Passed: false, Producer: producerName, Consumer: consumerName, Stage: "adjudicate", Step: "reject-mismatch"}
}
