package couplingexplain

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validStableID(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := semantic.ParseIdentity(value)
	return err == nil && parsed.String() == value
}
func missingIssue() *envelopeIssue {
	return &envelopeIssue{status: StatusUnknown, reason: ReasonMissing, code: "missing-verified-material"}
}
func validateCanonicalDigests(envelope VerifiedEnvelope) *envelopeIssue {
	if envelope.CodeBinding.CodeBindingDigest != codeBindingDigest(envelope.CodeBinding) {
		return integrityIssue("code-binding-digest-mismatch")
	}
	if envelope.Term.DefinitionDigest != termDefinitionDigest(envelope.Term) {
		return integrityIssue("term-digest-mismatch")
	}
	if envelope.OriginPath.PathDigest != pathDigest(envelope.OriginPath) {
		return integrityIssue("path-digest-mismatch", envelope.OriginPath.PathID)
	}
	if envelope.Receipt.ReceiptDigest != receiptDigest(envelope.Receipt) {
		return integrityIssue("receipt-digest-mismatch", envelope.Receipt.ReceiptID)
	}
	if envelope.Verifier.VerifierDigest != verifierDigest(envelope.Verifier) {
		return integrityIssue("verifier-digest-mismatch", envelope.Verifier.EvidenceID)
	}
	if envelope.Receipt.ChangeClaim == ClaimDelta &&
		(envelope.Receipt.CanonicalDelta == "" || envelope.Receipt.DeltaDigest != DigestBytes([]byte(envelope.Receipt.CanonicalDelta))) {
		return integrityIssue("delta-digest-mismatch", envelope.Receipt.ReceiptID)
	}
	if envelope.Receipt.ChangeClaim == ClaimNoDelta &&
		(envelope.Receipt.CanonicalDelta != "" || envelope.Receipt.DeltaDigest != "") {
		return integrityIssue("no-delta-carries-delta", envelope.Receipt.ReceiptID)
	}
	return nil
}
func validateEvidenceChain(envelope VerifiedEnvelope) *envelopeIssue {
	pathRefs := make([]string, 0, len(envelope.OriginPath.Steps))
	for _, step := range envelope.OriginPath.Steps {
		if step.EvidenceRef != "" {
			pathRefs = append(pathRefs, step.EvidenceRef)
		}
	}
	if len(pathRefs) != 1 || !sameUniqueStrings(pathRefs, envelope.Receipt.EvidenceRefs) {
		return integrityIssue("evidence-chain-mismatch", envelope.Receipt.ReceiptID)
	}
	if envelope.Verifier.EvidenceID != pathRefs[0] || envelope.Verifier.EvidenceID == "" {
		return integrityIssue("verifier-evidence-id-mismatch", envelope.Verifier.EvidenceID)
	}
	if !sameUniqueStrings(envelope.Verifier.EvidenceRefs, []string{envelope.OriginPath.PathID}) {
		return integrityIssue("verifier-path-evidence-mismatch", envelope.Verifier.EvidenceID)
	}
	return nil
}
func sameUniqueStrings(left, right []string) bool {
	left = sortedStrings(left)
	right = sortedStrings(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] == "" || (index > 0 && left[index] == left[index-1]) || left[index] != right[index] {
			return false
		}
	}
	return true
}
