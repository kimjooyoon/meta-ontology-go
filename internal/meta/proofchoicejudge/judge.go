package proofchoicejudge

import (
	"encoding/json"
	"reflect"
)

func Judge(data []byte) Verdict {
	result := Verdict{Schema: "gooo/proof-choice-algebra-judge/v3", ReceiptOnly: true, IndependentEvidence: 0, Decision: "FAIL_CLOSED", Reason: "RECEIPT_ONLY_NO_INDEPENDENT_EVIDENCE"}
	var input receipt
	if json.Unmarshal(data, &input) == nil {
		result.ReceiptDigest, result.Items, result.Transitions = input.Digest, len(input.Items), len(input.Transitions)
	}
	return result
}

func JudgeSource(path string, source, receiptData, before, after, baseline []byte) Verdict {
	result := Verdict{Schema: "gooo/proof-choice-algebra-judge/v3", IndependentEvidence: 1}
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
	expected := resolve(lowered.Values, lowered, baseline)
	result.EffectsMatch = reflect.DeepEqual(input.Effects, effectsFor(before, after))
	match := result.ReceiptDigestMatch && result.SourceDigestMatch && result.SemanticDigestMatch && result.EffectsMatch
	match = match && input.Schema == "gooo/proof-choice-algebra-receipt/v3" && input.SourcePath == path
	match = match && input.SourceReconstruction == result.SourceReconstruction
	match = match && input.Decision == expected.Decision && input.Reason == expected.Reason && input.SubjectResolution == expected.Resolution
	match = match && reflect.DeepEqual(input.Items, expected.Items) && reflect.DeepEqual(input.Evidence, expected.Evidence)
	match = match && reflect.DeepEqual(input.Compositions, expected.Compositions)
	match = match && reflect.DeepEqual(input.Transitions, expected.Transitions) && reflect.DeepEqual(input.Summary, expected.Summary)
	if !match {
		return fail(result, "INDEPENDENT_SOURCE_MISMATCH")
	}
	result.Decision, result.Reason = expected.Decision, expected.Reason
	return result
}

func fail(result Verdict, reason string) Verdict {
	result.Decision, result.Reason = "FAIL_CLOSED", reason
	return result
}
