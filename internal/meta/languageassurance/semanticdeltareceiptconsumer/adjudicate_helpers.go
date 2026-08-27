package semanticdeltareceiptconsumer

import "reflect"

func classDecision(structural StructuralDelta, claims ClaimDelta) (string, string, string) {
	return classDecisionWithComponents(structural, SemanticComponentDelta{}, claims)
}

func classDecisionWithComponents(structural StructuralDelta, components SemanticComponentDelta, claims ClaimDelta) (string, string, string) {
	if hasSemanticDelta(structural, components, claims) {
		if len(components.Added)+len(components.Removed)+len(components.Changed) > 0 && len(structural.AddedNodes)+len(structural.RemovedNodes)+len(structural.AddedFacts)+len(structural.RemovedFacts)+len(claims.Added)+len(claims.Removed)+len(claims.Changed) == 0 {
			return classChanged, decisionDelta, reasonComponentDelta
		}
		return classChanged, decisionDelta, reasonMeaning
	}
	return classPreserved, decisionFixedPoint, reasonTextualOnly
}

func semanticDecision(class string) string {
	if class == classPreserved {
		return semanticPreserved
	}
	return semanticChanged
}

func adjudicateUnavailable(base Verdict, receipt Receipt, before, after projectedSource, beforeRaw, afterRaw []byte, beforeErr, afterErr error) Verdict {
	base.Decision, base.Resolution, base.Classification, base.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonUnavailable
	base.Stage, base.Step = unavailableStage, unavailableStep
	expectedLedger, expectedTransitions := unknownLedger(before, after, unavailableStage, unavailableStep, reasonUnavailable)
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == base.Resolution && receipt.Classification == classIndeterminate && receipt.Reason == base.Reason && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && receipt.Stage == unavailableStage && receipt.Step == unavailableStep && receipt.StructuralDelta.Status == "UNKNOWN" && receipt.SemanticComponentDelta.Status == "UNKNOWN" && receipt.SemanticClaimDelta.Status == "UNKNOWN" && reflect.DeepEqual(receipt.Before, snapshot(beforeRaw, before, beforeErr)) && reflect.DeepEqual(receipt.After, snapshot(afterRaw, after, afterErr)) && reflect.DeepEqual(receipt.ClaimLedger, expectedLedger) && reflect.DeepEqual(receipt.ClaimTransitions, expectedTransitions)
	return base
}

func adjudicateSubjectUnknown(base Verdict, receipt Receipt, before, after projectedSource, beforeRaw, afterRaw []byte) Verdict {
	base.Decision, base.Resolution, base.Classification, base.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonSubject
	base.Step = "replay-unknown-subject"
	expectedLedger, expectedTransitions := unknownLedger(before, after, "bind-subject", "resolve-subject", reasonSubject)
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == resolutionLower && receipt.Classification == classIndeterminate && receipt.Reason == reasonSubject && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && receipt.Stage == "bind-subject" && receipt.Step == "resolve-subject" && receipt.StructuralDelta.Status == "UNKNOWN" && receipt.SemanticComponentDelta.Status == "UNKNOWN" && receipt.SemanticClaimDelta.Status == "UNKNOWN" && reflect.DeepEqual(receipt.ClaimLedger, expectedLedger) && reflect.DeepEqual(receipt.ClaimTransitions, expectedTransitions)
	base.Passed = base.Passed && reflect.DeepEqual(receipt.Before, snapshot(beforeRaw, before, nil)) && reflect.DeepEqual(receipt.After, snapshot(afterRaw, after, nil))
	return base
}

func adjudicateAmbiguous(base Verdict, receipt Receipt, before, after projectedSource, beforeRaw, afterRaw []byte, structural StructuralDelta, components SemanticComponentDelta, claims ClaimDelta) Verdict {
	base.Decision, base.Resolution, base.Classification, base.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonAmbiguous
	base.Stage, base.Step = "claim-delta", "match-claims"
	ledger, transitions := unknownLedger(before, after, base.Stage, base.Step, base.Reason)
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == base.Resolution && receipt.Classification == base.Classification && receipt.Reason == base.Reason && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && reflect.DeepEqual(receipt.StructuralDelta, structural) && reflect.DeepEqual(receipt.SemanticComponentDelta, components) && reflect.DeepEqual(receipt.SemanticClaimDelta, claims) && reflect.DeepEqual(receipt.ClaimLedger, ledger) && reflect.DeepEqual(receipt.ClaimTransitions, transitions)
	return base
}

func adjudicateUnmodeled(base Verdict, receipt Receipt, before, after projectedSource, beforeRaw, afterRaw []byte, structural StructuralDelta) Verdict {
	base.Decision, base.Resolution, base.Classification, base.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonUnmodeled
	base.Stage, base.Step = "semantic-projection", "compare-stable-hash"
	ledger, transitions := unknownLedger(before, after, base.Stage, base.Step, base.Reason)
	components := SemanticComponentDelta{Status: "UNKNOWN"}
	claims := ClaimDelta{Status: "UNKNOWN", Reason: reasonUnmodeled}
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == base.Resolution && receipt.Classification == base.Classification && receipt.Reason == base.Reason && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && reflect.DeepEqual(receipt.StructuralDelta, structural) && reflect.DeepEqual(receipt.SemanticComponentDelta, components) && reflect.DeepEqual(receipt.SemanticClaimDelta, claims) && reflect.DeepEqual(receipt.ClaimLedger, ledger) && reflect.DeepEqual(receipt.ClaimTransitions, transitions)
	return base
}

func mismatchVerdict() Verdict {
	return Verdict{Decision: decisionFailClosed, Resolution: resolutionInvariant, Classification: classIndeterminate, Reason: reasonReceipt, Passed: false, Producer: producerName, Consumer: consumerName, Stage: "adjudicate", Step: "reject-mismatch"}
}
