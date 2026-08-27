package semanticdeltareceiptconsumer

import (
	"os"
	"reflect"
)

// AdjudicateFiles is a separate consumer. It reads the files, rebuilds every
// projection and validates the receipt digest before making a decision.
func AdjudicateFiles(input Input, receipt Receipt) Verdict {
	base := Verdict{Producer: receipt.Producer, Consumer: consumerName, Stage: "adjudicate", Step: "replay-receipt"}
	beforeRaw, beforeReadErr := os.ReadFile(input.BeforePath)
	afterRaw, afterReadErr := os.ReadFile(input.AfterPath)
	if beforeReadErr != nil || afterReadErr != nil {
		return mismatchVerdict()
	}
	before, beforeErr := projectSourceSide(input.BeforePath, beforeRaw, true)
	after, afterErr := projectSourceSide(input.AfterPath, afterRaw, false)
	if before.path == "" {
		before = sourceEnvelope(input.BeforePath, beforeRaw)
	}
	if after.path == "" {
		after = sourceEnvelope(input.AfterPath, afterRaw)
	}
	meta, metaErr := readMetaContract()
	text := textualDelta(beforeRaw, afterRaw)
	if !receiptIdentityMatches(input, receipt, beforeRaw, afterRaw, text, meta) {
		return mismatchVerdict()
	}
	if metaErr != nil {
		return adjudicateUnavailable(base, receipt, before, after, beforeRaw, afterRaw, metaErr, nil)
	}
	if beforeErr != nil || afterErr != nil {
		return adjudicateUnavailable(base, receipt, before, after, beforeRaw, afterRaw, beforeErr, afterErr)
	}
	if !validSubject(input.SubjectSHA) {
		return adjudicateSubjectUnknown(base, receipt, before, after, beforeRaw, afterRaw)
	}
	structural := structuralDelta(before, after)
	components := semanticComponentDelta(before, after)
	claims := claimDelta(before, after)
	if claims.Status == "UNKNOWN" {
		return adjudicateAmbiguous(base, receipt, before, after, beforeRaw, afterRaw, structural, components, claims)
	}
	if before.semanticDigest != after.semanticDigest && len(components.Added)+len(components.Removed)+len(components.Changed) == 0 {
		return adjudicateUnmodeled(base, receipt, before, after, beforeRaw, afterRaw, structural)
	}
	class, decision, reason := classDecisionWithComponents(structural, components, claims)
	expectedLedger, expected := claimLedger(before, after, class, reason)
	base.Decision, base.Resolution, base.Classification, base.Reason = decision, resolutionExact, class, reason
	expectedBefore, expectedAfter := snapshot(beforeRaw, before, nil), snapshot(afterRaw, after, nil)
	base.Passed = receipt.Decision == decision && receipt.Resolution == resolutionExact && receipt.Classification == class && receipt.Reason == reason && receipt.RawDecision == text.Decision && receipt.SemanticDecision == semanticDecision(class) && reflect.DeepEqual(receipt.Before, expectedBefore) && reflect.DeepEqual(receipt.After, expectedAfter) && reflect.DeepEqual(receipt.StructuralDelta, structural) && reflect.DeepEqual(receipt.SemanticComponentDelta, components) && reflect.DeepEqual(receipt.SemanticClaimDelta, claims) && reflect.DeepEqual(receipt.ClaimLedger, expectedLedger) && reflect.DeepEqual(receipt.ClaimTransitions, expected) && receipt.MetaContractDigest == meta.Digest && receipt.DenominatorVersion == meta.Version && receipt.DenominatorCases == meta.DenominatorCases && receipt.ModeledSemanticComponents == modeledComponentCount && receipt.TotalSemanticComponents == totalComponentCount && receipt.SemanticCoverageBPS == 10000 && receipt.TransitionCount == len(expected)
	return base
}

func receiptIdentityMatches(input Input, receipt Receipt, before, after []byte, text TextualDelta, meta metaContract) bool {
	return receipt.Schema == receiptSchema && receipt.Producer == producerName && receipt.Consumer == consumerName && receipt.MetaOperation == metaOperation && receipt.CaseID == input.CaseID && receipt.SubjectSHA == input.SubjectSHA && receipt.Before.SourceDigest == digestBytes(before) && receipt.After.SourceDigest == digestBytes(after) && reflect.DeepEqual(receipt.TextualDelta, text) && (receipt.MetaSourcePath == "" || receipt.MetaSourcePath == metaSourcePath) && (meta.Digest == "" || receipt.MetaContractDigest == "" || receipt.MetaContractDigest == meta.Digest) && receiptDigestValid(receipt)
}
