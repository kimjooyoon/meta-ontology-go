package couplingexplain

func validateRequestBinding(request Request, envelope VerifiedEnvelope) *envelopeIssue {
	if request.CodeSymbolID == "" || request.SnapshotDigest == "" || request.RegistryDigest == "" ||
		request.SourceMapDigest == "" || request.ManifestDigest == "" || request.ToolchainDigest == "" || request.ProfileDigest == "" ||
		request.DetectorInputDigest == "" || request.DetectorResultDigest == "" || request.VerifierResultDigest == "" {
		return &envelopeIssue{status: StatusUnknown, reason: ReasonMissing, code: "missing-request-binding"}
	}
	if !validStableID(request.CodeSymbolID) || !validDigest(request.SnapshotDigest) || !validDigest(request.RegistryDigest) ||
		!validDigest(request.SourceMapDigest) || !validDigest(request.ManifestDigest) || !validDigest(request.ToolchainDigest) || !validDigest(request.ProfileDigest) ||
		!validDigest(request.DetectorInputDigest) || !validDigest(request.DetectorResultDigest) || !validDigest(request.VerifierResultDigest) {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonStale, code: "malformed-request-binding"}
	}
	if request.EnvelopeDigest != "" && !validDigest(request.EnvelopeDigest) {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonStale, code: "malformed-request-envelope-digest"}
	}
	if envelope.EnvelopeDigest == "" || request.EnvelopeDigest != envelope.EnvelopeDigest {
		return &envelopeIssue{status: StatusUnknown, reason: ReasonStale, code: "stale-envelope-digest"}
	}
	if envelope.Digest() != envelope.EnvelopeDigest {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonStale, code: "envelope-digest-mismatch"}
	}
	expected := SnapshotBinding{SnapshotDigest: request.SnapshotDigest, RegistryDigest: request.RegistryDigest,
		SourceMapDigest: request.SourceMapDigest, ManifestDigest: request.ManifestDigest, ToolchainDigest: request.ToolchainDigest, ProfileDigest: request.ProfileDigest,
		DetectorInputDigest: request.DetectorInputDigest, DetectorResultDigest: request.DetectorResultDigest, VerifierResultDigest: request.VerifierResultDigest, Control: request.Control}
	if envelope.Binding != expected {
		return &envelopeIssue{status: StatusUnknown, reason: ReasonStale, code: "stale-snapshot-binding"}
	}
	if envelope.CodeBinding.CodeSymbolID == "" && envelope.Verdict != VerdictVerified {
		return nil
	}
	if envelope.CodeBinding.CodeSymbolID != request.CodeSymbolID {
		return &envelopeIssue{status: StatusUnknown, reason: ReasonUnregistered, code: "code-symbol-unregistered", ids: []string{request.CodeSymbolID}}
	}
	if envelope.Upstream != nil && (envelope.Upstream.SourceMapDigest != "" && envelope.Upstream.SourceMapDigest != request.SourceMapDigest ||
		envelope.Upstream.ManifestDigest != request.ManifestDigest || envelope.Upstream.DetectorInputDigest != request.DetectorInputDigest ||
		envelope.Upstream.DetectorResultDigest != request.DetectorResultDigest) {
		return &envelopeIssue{status: StatusUnknown, reason: ReasonStale, code: "stale-upstream-digest"}
	}
	return nil
}
