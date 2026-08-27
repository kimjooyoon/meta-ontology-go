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
	before, beforeErr := projectSource(input.BeforePath, beforeRaw)
	after, afterErr := projectSource(input.AfterPath, afterRaw)
	text := textualDelta(beforeRaw, afterRaw)
	if !receiptIdentityMatches(input, receipt, beforeRaw, afterRaw, text) {
		return mismatchVerdict()
	}
	if !validSubject(input.SubjectSHA) {
		return adjudicateSubjectUnknown(base, receipt, before, after, beforeRaw, afterRaw)
	}
	if beforeErr != nil || afterErr != nil {
		return adjudicateUnavailable(base, receipt, before, after, beforeRaw, afterRaw, beforeErr, afterErr)
	}
	structural := structuralDelta(before, after)
	claims := claimDelta(before, after)
	class, decision, reason := classDecision(structural, claims)
	expected := transitions(before, after, class, reason)
	base.Decision, base.Resolution, base.Classification, base.Reason = decision, resolutionExact, class, reason
	base.Passed = receipt.Decision == decision && receipt.Resolution == resolutionExact && receipt.Classification == class && receipt.Reason == reason && receipt.RawDecision == text.Decision && receipt.SemanticDecision == semanticDecision(class) && reflect.DeepEqual(receipt.Before, snapshot(beforeRaw, before, nil)) && reflect.DeepEqual(receipt.After, snapshot(afterRaw, after, nil)) && reflect.DeepEqual(receipt.StructuralDelta, structural) && reflect.DeepEqual(receipt.SemanticClaimDelta, claims) && reflect.DeepEqual(receipt.ClaimTransitions, expected)
	return base
}

func receiptIdentityMatches(input Input, receipt Receipt, before, after []byte, text TextualDelta) bool {
	return receipt.Schema == receiptSchema && receipt.Producer == producerName && receipt.Consumer == consumerName && receipt.MetaOperation == metaOperation && receipt.CaseID == input.CaseID && receipt.SubjectSHA == input.SubjectSHA && receipt.Before.SourceDigest == digestBytes(before) && receipt.After.SourceDigest == digestBytes(after) && reflect.DeepEqual(receipt.TextualDelta, text) && receipt.RepositoryWrites == 0 && receiptDigestValid(receipt)
}
