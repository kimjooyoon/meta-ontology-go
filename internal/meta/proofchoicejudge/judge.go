package proofchoicejudge

import "encoding/json"

func Judge(data []byte) Verdict {
	var input receipt
	result := Verdict{Schema: "gooo/proof-choice-algebra-judge/v1", Independent: true}
	if err := json.Unmarshal(data, &input); err != nil {
		result.Decision, result.Reason = failClosed, "RECEIPT_JSON_UNKNOWN"
		return result
	}
	result.ReceiptDigest = input.Digest
	result.Items, result.Transitions = len(input.Items), len(input.Transitions)
	computed, err := digest(input)
	if err != nil {
		result.Decision, result.Reason = failClosed, "RECEIPT_DIGEST_UNKNOWN"
		return result
	}
	result.ComputedDigest, result.DigestMatch = computed, computed == input.Digest
	reason := validate(input)
	want := pass
	if reason != "" {
		want = failClosed
	}
	if !result.DigestMatch {
		want, reason = failClosed, "RECEIPT_DIGEST_MISMATCH"
	}
	if input.Decision != want && reason == "" {
		want, reason = failClosed, "PRODUCER_DECISION_MISMATCH"
	}
	result.Decision, result.Reason = want, reason
	if result.Decision == pass {
		result.Reason = "PROOF_CHOICES_COMPOSED"
	}
	return result
}
