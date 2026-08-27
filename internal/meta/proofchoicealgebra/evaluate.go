package proofchoicealgebra

import "fmt"

func Evaluate(path string, source, before, after, baseline []byte) Receipt {
	lowered, lowerErr := lowerSource(path, source)
	receipt := Receipt{
		Schema: ReceiptSchema, Decision: Pass, Reason: "PROOF_VALUES_RESOLVED", SubjectResolution: Exact,
		SourcePath: path, SourceDigest: digestBytes(source), SemanticDigest: lowered.SemanticDigest,
		SourceReconstruction: Reconstruction{lowered.Reconstructed, lowered.ReconstructionDenom},
		Summary:              baseSummary(),
		Effects:              ObserveEffects(before, after),
	}
	if lowerErr != nil {
		receipt.Decision, receipt.Reason, receipt.SubjectResolution = FailClosed, lowerErr.Error(), Lower
		return sealOrFallback(receipt)
	}
	items, evidence, compositions, transitions, summary, decision, reason, resolution := resolve(lowered.Values, lowered, baseline)
	receipt.Items, receipt.Evidence, receipt.Compositions = items, evidence, compositions
	receipt.Transitions, receipt.Summary = transitions, summary
	receipt.Decision, receipt.Reason, receipt.SubjectResolution = decision, reason, resolution
	return sealOrFallback(receipt)
}

func sealOrFallback(receipt Receipt) Receipt {
	sealed, err := Seal(receipt)
	if err != nil {
		receipt.Decision, receipt.Reason, receipt.SubjectResolution = FailClosed, fmt.Sprintf("RECEIPT_DIGEST_UNKNOWN: %v", err), Lower
		return receipt
	}
	return sealed
}
