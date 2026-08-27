package semanticdeltareceiptconsumer

import "reflect"

func classDecision(structural StructuralDelta, claims ClaimDelta) (string, string, string) {
	if hasSemanticDelta(structural, claims) {
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
	expectedLedger, expectedTransitions := unknownLedger(unavailableStage, unavailableStep, reasonUnavailable)
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == base.Resolution && receipt.Classification == classIndeterminate && receipt.Reason == base.Reason && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && receipt.Stage == unavailableStage && receipt.Step == unavailableStep && receipt.StructuralDelta.Status == "UNKNOWN" && receipt.SemanticClaimDelta.Status == "UNKNOWN" && reflect.DeepEqual(receipt.Before, snapshot(beforeRaw, before, beforeErr)) && reflect.DeepEqual(receipt.After, snapshot(afterRaw, after, afterErr)) && reflect.DeepEqual(receipt.ClaimLedger, expectedLedger) && reflect.DeepEqual(receipt.ClaimTransitions, expectedTransitions)
	return base
}

func adjudicateSubjectUnknown(base Verdict, receipt Receipt, before, after projectedSource, beforeRaw, afterRaw []byte) Verdict {
	base.Decision, base.Resolution, base.Classification, base.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonSubject
	base.Step = "replay-unknown-subject"
	expectedLedger, expectedTransitions := unknownLedger("bind-subject", "resolve-subject", reasonSubject)
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == resolutionLower && receipt.Classification == classIndeterminate && receipt.Reason == reasonSubject && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && receipt.Stage == "bind-subject" && receipt.Step == "resolve-subject" && receipt.StructuralDelta.Status == "UNKNOWN" && receipt.SemanticClaimDelta.Status == "UNKNOWN" && reflect.DeepEqual(receipt.ClaimLedger, expectedLedger) && reflect.DeepEqual(receipt.ClaimTransitions, expectedTransitions)
	base.Passed = base.Passed && reflect.DeepEqual(receipt.Before, snapshot(beforeRaw, before, nil)) && reflect.DeepEqual(receipt.After, snapshot(afterRaw, after, nil))
	return base
}

func mismatchVerdict() Verdict {
	return Verdict{Decision: decisionFailClosed, Resolution: resolutionInvariant, Classification: classIndeterminate, Reason: reasonReceipt, Passed: false, Producer: producerName, Consumer: consumerName, Stage: "adjudicate", Step: "reject-mismatch"}
}
