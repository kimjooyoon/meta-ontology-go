package proofchoicejudge

import (
	"encoding/json"
	"reflect"
)

func Judge(data []byte) Verdict {
	result := Verdict{Schema: "gooo/proof-choice-algebra-judge/v2", ReceiptOnly: true, IndependentEvidence: 0, Decision: "FAIL_CLOSED", Reason: "RECEIPT_ONLY_NO_INDEPENDENT_EVIDENCE"}
	var input receipt
	if json.Unmarshal(data, &input) == nil {
		result.ReceiptDigest, result.Items, result.Transitions = input.Digest, len(input.Items), len(input.Transitions)
	}
	return result
}

func JudgeSource(path string, source, receiptData, before, after []byte) Verdict {
	result := Verdict{Schema: "gooo/proof-choice-algebra-judge/v2", IndependentEvidence: 1}
	var input receipt
	if err := json.Unmarshal(receiptData, &input); err != nil {
		return fail(result, "RECEIPT_JSON_UNKNOWN")
	}
	result.ReceiptDigest, result.Items, result.Transitions = input.Digest, len(input.Items), len(input.Transitions)
	computed, err := digestReceipt(input)
	if err != nil {
		return fail(result, "RECEIPT_DIGEST_UNKNOWN")
	}
	result.ComputedReceiptDigest, result.ReceiptDigestMatch = computed, computed == input.Digest
	lowered, lowerErr := lowerSource(path, source)
	result.SourceReconstruction = reconstruction{lowered.Reconstructed, lowered.ReconstructionDenom}
	result.SourceDigest, result.ComputedSourceDigest = input.SourceDigest, digestBytes(source)
	result.SourceDigestMatch = result.SourceDigest == result.ComputedSourceDigest
	result.SemanticDigest, result.ComputedSemanticDigest = input.SemanticDigest, lowered.SemanticDigest
	result.SemanticDigestMatch = result.SemanticDigest == result.ComputedSemanticDigest
	if lowerErr != nil {
		return fail(result, lowerErr.Error())
	}
	expected := resolve(lowered.Values)
	result.EffectsMatch = reflect.DeepEqual(input.Effects, effectsFor(before, after))
	if !result.ReceiptDigestMatch || !result.SourceDigestMatch || !result.SemanticDigestMatch || !result.EffectsMatch || input.SourcePath != path || input.SourceReconstruction != result.SourceReconstruction || input.Decision != expected.Decision || input.Reason != expected.Reason || input.SubjectResolution != expected.Resolution || !reflect.DeepEqual(input.Items, expected.Items) || !reflect.DeepEqual(input.Transitions, expected.Transitions) || !reflect.DeepEqual(input.Summary, expected.Summary) {
		return fail(result, "INDEPENDENT_SOURCE_MISMATCH")
	}
	result.Decision, result.Reason = expected.Decision, expected.Reason
	return result
}

func fail(result Verdict, reason string) Verdict {
	result.Decision, result.Reason = "FAIL_CLOSED", reason
	return result
}
