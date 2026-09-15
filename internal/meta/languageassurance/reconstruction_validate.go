package languageassurance

import (
	"encoding/hex"
	"slices"
	"strings"
)

func validRawReconstructions(receipts []RawReconstructionReceipt) bool {
	if len(receipts) > 1 {
		return false
	}
	for _, receipt := range receipts {
		observation := receipt.Observation
		if receipt.Schema != RawReconstructionSchema || receipt.VerifierID != RawVerifierID || !validSHA(receipt.SubjectSHA) {
			return false
		}
		if !validSHA256(receipt.DenominatorDigest) || !validSHA256(receipt.RawTransactionDigest) {
			return false
		}
		if !slices.Contains([]string{CandidateAllowLimited, CandidateBlock, CandidateFailClosed}, observation.CandidateDecision) || observation.CandidateReason == "" {
			return false
		}
		if !slices.Contains([]Resolution{ResolutionExact, ResolutionInvariantOnly, ResolutionUnknown}, observation.CandidateResolution) {
			return false
		}
	}
	return true
}

func validSHA256(value string) bool {
	raw, ok := strings.CutPrefix(value, "sha256:")
	if !ok || len(raw) != 64 {
		return false
	}
	_, err := hex.DecodeString(raw)
	return err == nil
}
