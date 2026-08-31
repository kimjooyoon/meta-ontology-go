package selfimprovementattestation

import "fmt"

func Resolve(request Request) (ResolutionReceipt, error) {
	receipt := baseReceipt(request)
	if err := validatePrior(request); err != nil {
		return failClosed(receipt, "PRIOR_TRANSPORT_RECEIPT_INVALID", err)
	}
	if err := validateArchive(request); err != nil {
		return failClosed(receipt, err.Error(), err)
	}
	if request.VerifierExitCode != 0 {
		receipt.Decision = "UNKNOWN"
		receipt.Resolution = "LOWER_RESOLUTION"
		receipt.Reason = "PRODUCER_ATTESTATION_UNAVAILABLE"
		receipt.Coordinate = Coordinate{Stage: "ATTEST", Step: "cryptographic-verify"}
		receipt.Views = lowerResolutionViews()
		receipt.Proofs = baselineProofs(receipt)
		if err := seal(&receipt); err != nil {
			return receipt, err
		}
		return receipt, nil
	}
	result, mismatches := selectVerification(request)
	if result == nil {
		reason := mismatches[0]
		return failClosed(receipt, reason, fmt.Errorf("%s", reason))
	}
	evidence, err := digestValue(result)
	if err != nil {
		return failClosed(receipt, "ATTESTATION_EVIDENCE_DIGEST_FAILED", err)
	}
	observe(&receipt, request, result, evidence)
	if err := seal(&receipt); err != nil {
		return receipt, err
	}
	return receipt, nil
}

func failClosed(receipt ResolutionReceipt, reason string, cause error) (ResolutionReceipt, error) {
	receipt.Decision = "FAIL_CLOSED"
	receipt.Resolution = "EXACT"
	receipt.Reason = reason
	receipt.Obligations = setAttestation(receipt.Obligations, "FALSE", reason, "")
	receipt.OpenObligationIDs = []string{attestationID}
	receipt.Metrics = Metrics{8, 7, 0, 1, 1, 8750, 0}
	receipt.Views = lowerResolutionViews()
	receipt.ClaimTransitions = []ClaimTransition{{ClaimID: attestationID, Before: "OPEN", After: "REJECTED"}}
	receipt.Proofs = baselineProofs(receipt)
	if err := seal(&receipt); err != nil {
		return receipt, err
	}
	return receipt, cause
}
