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
	base.Decision, base.Resolution, base.Classification, base.Reason = decisionFailClosed, resolutionUnknown, classIndeterminate, reasonUnavailable
	base.Step = "replay-unknown-source"
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == base.Resolution && receipt.Classification == base.Classification && receipt.Reason == base.Reason && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && receipt.StructuralDelta.Status == "UNKNOWN" && receipt.SemanticClaimDelta.Status == "UNKNOWN" && reflect.DeepEqual(receipt.Before, snapshot(beforeRaw, before, beforeErr)) && reflect.DeepEqual(receipt.After, snapshot(afterRaw, after, afterErr)) && reflect.DeepEqual(receipt.ClaimTransitions, transitions(before, after, classIndeterminate, reasonUnavailable))
	_ = beforeErr
	_ = afterErr
	return base
}

func adjudicateSubjectUnknown(base Verdict, receipt Receipt, before, after projectedSource, beforeRaw, afterRaw []byte) Verdict {
	base.Decision, base.Resolution, base.Classification, base.Reason = decisionFailClosed, resolutionLower, classIndeterminate, reasonSubject
	base.Step = "replay-unknown-subject"
	base.Passed = receipt.Decision == base.Decision && receipt.Resolution == base.Resolution && receipt.Classification == base.Classification && receipt.Reason == base.Reason && receipt.RawDecision == textualDelta(beforeRaw, afterRaw).Decision && receipt.SemanticDecision == semanticUnknown && receipt.Stage == "bind-subject" && receipt.Step == "resolve-subject" && receipt.StructuralDelta.Status == "UNKNOWN" && receipt.SemanticClaimDelta.Status == "UNKNOWN" && reflect.DeepEqual(receipt.ClaimTransitions, []ClaimTransition{{ClaimID: boundedClaimID, FromStatus: statusOpen, ToStatus: statusOpen, Stage: "bind-subject", Step: "resolve-subject", Reason: reasonSubject}})
	base.Passed = base.Passed && reflect.DeepEqual(receipt.Before, snapshot(beforeRaw, before, nil)) && reflect.DeepEqual(receipt.After, snapshot(afterRaw, after, nil))
	return base
}

func mismatchVerdict() Verdict {
	return Verdict{Decision: decisionFailClosed, Resolution: resolutionInvariant, Classification: classIndeterminate, Reason: reasonReceipt, Passed: false, Producer: producerName, Consumer: consumerName, Stage: "adjudicate", Step: "reject-mismatch"}
}
