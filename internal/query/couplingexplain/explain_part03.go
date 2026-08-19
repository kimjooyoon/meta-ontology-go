package couplingexplain

func validateEnvelopeShape(envelope VerifiedEnvelope) *envelopeIssue {
	if envelope.Schema == "" || envelope.Binding.SnapshotDigest == "" || !validDigest(envelope.EvidenceDigest) {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonMissing, code: "incomplete-envelope-header"}
	}
	if envelope.Schema != "gooo-coupling-explanation/v1" {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonNotVerified, code: "malformed-envelope-schema"}
	}
	if envelope.Upstream != nil {
		if (envelope.Upstream.SourceMapDigest != "" && !validDigest(envelope.Upstream.SourceMapDigest)) || !validDigest(envelope.Upstream.ManifestDigest) || !validDigest(envelope.Upstream.DetectorInputDigest) ||
			!validDigest(envelope.Upstream.DetectorResultDigest) || (envelope.Upstream.DetectorStatus != "" &&
			envelope.Upstream.DetectorStatus != "PASS" && envelope.Upstream.DetectorStatus != "UNKNOWN" && envelope.Upstream.DetectorStatus != "FAIL_CLOSED") {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonStale, code: "malformed-upstream-evidence"}
		}
	}
	if envelope.Verdict != VerdictVerified && envelope.Verdict != VerdictUnknown && envelope.Verdict != VerdictFailClosed {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonStale, code: "invalid-envelope-verdict"}
	}
	if envelope.Verdict != VerdictVerified {
		if !validReason(envelope.NoLinkReason) || len(envelope.Diagnostics) == 0 {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonMissing, code: "incomplete-no-link-envelope"}
		}
		if (envelope.Receipt.ChangeClaim != "" && !validChangeClaim(envelope.Receipt.ChangeClaim)) ||
			(envelope.Receipt.ReceiptKind != "" && !envelope.Receipt.ReceiptKind.Valid()) ||
			(envelope.Verifier.State != "" && envelope.Verifier.State != VerifierPass &&
				envelope.Verifier.State != VerifierFailClosed && envelope.Verifier.State != VerifierUnknown) ||
			(envelope.OriginPath.StepCount != 0 && envelope.OriginPath.StepCount != len(envelope.OriginPath.Steps)) {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonStale, code: "contradictory-no-link-envelope"}
		}
		return nil
	}
	if missingVerifiedMaterial(envelope) {
		return missingIssue()
	}
	if issue := malformedVerifiedMaterial(envelope); issue != nil {
		return issue
	}
	if envelope.SemanticOwner != envelope.CodeBinding.SemanticOwnerID || envelope.Term.SemanticOwnerID != envelope.SemanticOwner {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonUnregistered, code: "envelope-owner-mismatch"}
	}
	if envelope.NoLinkReason != "" || envelope.Verifier.State != VerifierPass || !envelope.Verifier.Independent ||
		envelope.Verifier.EvidenceDigest != envelope.EvidenceDigest || envelope.Verifier.ReceiptID != envelope.Receipt.ReceiptID ||
		envelope.Receipt.OriginPathID != envelope.OriginPath.PathID || envelope.OriginPath.StartID != envelope.CodeBinding.CodeSymbolID ||
		envelope.OriginPath.EndID == "" {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonNotVerified, code: "verified-envelope-contradiction"}
	}
	if inputIssue := validateConnectedPath(envelope.OriginPath, envelope.SemanticOwner, envelope.Term.TermID); inputIssue != nil {
		return inputIssue
	}
	if inputIssue := validateVerifiedIntegrity(envelope); inputIssue != nil {
		return inputIssue
	}
	return nil
}
