package proofchoicealgebra

import "fmt"

func Evaluate(path string, source, before, after []byte) Receipt {
	lowered, lowerErr := lowerSource(path, source)
	receipt := Receipt{
		Schema: ReceiptSchema, Decision: Pass, Reason: "PROOF_VALUES_RESOLVED", SubjectResolution: Exact,
		SourcePath: path, SourceDigest: digestBytes(source), SemanticDigest: lowered.SemanticDigest,
		SourceReconstruction: Reconstruction{lowered.Reconstructed, lowered.ReconstructionDenom},
		Summary:              Summary{FixedDenominator: FixedDenom, ChoiceCounts: map[Route]int{Foundation: 0, Coherence: 0, Regression: 0}},
		Effects:              ObserveEffects(before, after),
	}
	if lowerErr != nil {
		receipt.Decision, receipt.Reason, receipt.SubjectResolution = FailClosed, lowerErr.Error(), FailClosed
		return sealOrFallback(receipt)
	}
	items, transitions, summary, decision, reason, resolution := resolve(lowered.Values)
	receipt.Items, receipt.Transitions, receipt.Summary = items, transitions, summary
	receipt.Decision, receipt.Reason, receipt.SubjectResolution = decision, reason, resolution
	return sealOrFallback(receipt)
}

func sealOrFallback(receipt Receipt) Receipt {
	sealed, err := Seal(receipt)
	if err != nil {
		receipt.Decision, receipt.Reason, receipt.SubjectResolution = FailClosed, fmt.Sprintf("RECEIPT_DIGEST_UNKNOWN: %v", err), FailClosed
		return receipt
	}
	return sealed
}
