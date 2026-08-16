package couplingexplain

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

func validateVerifiedIntegrity(envelope VerifiedEnvelope) *envelopeIssue {
	if issue := validateCanonicalDigests(envelope); issue != nil {
		return issue
	}
	return validateEvidenceChain(envelope)
}

func missingVerifiedMaterial(envelope VerifiedEnvelope) bool {
	return envelope.CodeBinding.CodeSymbolID == "" || envelope.CodeBinding.SemanticOwnerID == "" ||
		envelope.CodeBinding.RegisteredSurfaceID == "" || envelope.CodeBinding.SourceMapID == "" ||
		envelope.CodeBinding.BindingDigest == "" || envelope.CodeBinding.CodeBindingDigest == "" ||
		envelope.SemanticOwner == "" || envelope.Term.TermID == "" || envelope.Term.SemanticOwnerID == "" ||
		envelope.Term.Version == "" || envelope.Term.DefinitionDigest == "" || envelope.OriginPath.PathID == "" ||
		envelope.OriginPath.StartID == "" || envelope.OriginPath.EndID == "" || envelope.OriginPath.StepCount < 1 ||
		len(envelope.OriginPath.Steps) == 0 || envelope.OriginPath.PathDigest == "" || envelope.Receipt.ReceiptID == "" ||
		envelope.Receipt.SurfaceID == "" || envelope.Receipt.ChangeClaim == "" || envelope.Receipt.ReceiptKind == "" ||
		envelope.Receipt.ReceiptDigest == "" || envelope.Receipt.OriginPathID == "" || envelope.Verifier.EvidenceID == "" ||
		envelope.Verifier.ReceiptID == "" || envelope.Verifier.State == "" || envelope.Verifier.EvidenceDigest == "" ||
		envelope.Verifier.VerifierDigest == "" || len(envelope.Verifier.EvidenceRefs) == 0 || len(envelope.Receipt.EvidenceRefs) == 0
}

func malformedVerifiedMaterial(envelope VerifiedEnvelope) *envelopeIssue {
	if !validStableID(envelope.CodeBinding.CodeSymbolID) || !validStableID(envelope.CodeBinding.SemanticOwnerID) ||
		!validStableID(envelope.CodeBinding.RegisteredSurfaceID) || !validStableID(envelope.CodeBinding.SourceMapID) ||
		!validDigest(envelope.CodeBinding.BindingDigest) || !validDigest(envelope.CodeBinding.CodeBindingDigest) ||
		!validStableID(envelope.SemanticOwner) || !validStableID(envelope.Term.TermID) || !validStableID(envelope.Term.SemanticOwnerID) ||
		!validDigest(envelope.Term.DefinitionDigest) || !validStableID(envelope.OriginPath.PathID) ||
		!validStableID(envelope.OriginPath.StartID) || !validStableID(envelope.OriginPath.EndID) ||
		!validDigest(envelope.OriginPath.PathDigest) || envelope.OriginPath.StepCount != len(envelope.OriginPath.Steps) ||
		!validStableID(envelope.Receipt.ReceiptID) || !validStableID(envelope.Receipt.SurfaceID) ||
		!validDigest(envelope.Receipt.BeforeIRDigest) || !validDigest(envelope.Receipt.AfterIRDigest) ||
		!validDigest(envelope.Receipt.ReceiptDigest) || !validStableID(envelope.Receipt.OriginPathID) ||
		!validStableID(envelope.Verifier.EvidenceID) || !validStableID(envelope.Verifier.ReceiptID) ||
		!validDigest(envelope.Verifier.EvidenceDigest) || !validDigest(envelope.Verifier.VerifierDigest) {
		return integrityIssue("malformed-verified-material")
	}
	if !validChangeClaim(envelope.Receipt.ChangeClaim) || !envelope.Receipt.ReceiptKind.Valid() {
		return integrityIssue("invalid-receipt-kind", envelope.Receipt.ReceiptID)
	}
	if (envelope.Receipt.ChangeClaim == ClaimDelta && envelope.Receipt.ReceiptKind != semantic.SemanticDelta) ||
		(envelope.Receipt.ChangeClaim == ClaimNoDelta && envelope.Receipt.ReceiptKind != semantic.NoSemanticDelta) {
		return integrityIssue("receipt-kind-mismatch", envelope.Receipt.ReceiptID)
	}
	if envelope.Receipt.DeltaDigest != "" && !validDigest(envelope.Receipt.DeltaDigest) {
		return integrityIssue("malformed-delta-digest", envelope.Receipt.ReceiptID)
	}
	if envelope.Verifier.State != VerifierPass && envelope.Verifier.State != VerifierFailClosed && envelope.Verifier.State != VerifierUnknown {
		return integrityIssue("invalid-verifier-state", envelope.Verifier.EvidenceID)
	}
	for _, step := range envelope.OriginPath.Steps {
		if !validStableID(step.FromID) || !validStableID(step.ToID) || !validDigest(step.InputDigest) || !validDigest(step.OutputDigest) ||
			(step.RuleRef != "" && !validStableID(step.RuleRef)) || (step.EvidenceRef != "" && !validStableID(step.EvidenceRef)) {
			return integrityIssue("malformed-origin-step", envelope.OriginPath.PathID)
		}
	}
	for _, evidenceRef := range envelope.Receipt.EvidenceRefs {
		if !validStableID(evidenceRef) {
			return integrityIssue("malformed-receipt-evidence", envelope.Receipt.ReceiptID)
		}
	}
	for _, evidenceRef := range envelope.Verifier.EvidenceRefs {
		if !validStableID(evidenceRef) {
			return integrityIssue("malformed-verifier-evidence", envelope.Verifier.EvidenceID)
		}
	}
	return nil
}

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

func integrityIssue(code string, ids ...string) *envelopeIssue {
	return &envelopeIssue{status: StatusFailClosed, reason: ReasonNotVerified, code: code, ids: ids}
}
