package couplingexplain

import (
	"context"
)

type envelopeIssue struct {
	status Status
	reason LinkReason
	code   string
	ids    []string
}

// Explain only consumes an already verified envelope. It checks exact input
// binding and output-shape integrity; it does not infer ownership, delta
// semantics, path meaning, or authorization.
func Explain(ctx context.Context, request Request, envelope VerifiedEnvelope) Explanation {
	binding := requestBinding(request)
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return requestNoLink(binding, StatusUnknown, ReasonStale, "cancelled")
	}
	if request.Control.RequestVersion != request.Control.ObservedVersion ||
		request.Control.RequestCancellationVersion != request.Control.ObservedCancellationVersion {
		return requestNoLink(binding, StatusUnknown, ReasonStale, "version-race")
	}
	if inputIssue := validateRequestBinding(request, envelope); inputIssue != nil {
		return requestNoLink(binding, inputIssue.status, inputIssue.reason, inputIssue.code, inputIssue.ids...)
	}
	if inputIssue := validateEnvelopeShape(envelope); inputIssue != nil {
		return requestNoLink(binding, inputIssue.status, inputIssue.reason, inputIssue.code, inputIssue.ids...)
	}
	if ctx.Err() != nil {
		return requestNoLink(binding, StatusUnknown, ReasonStale, "cancelled")
	}
	if envelope.Verdict != VerdictVerified {
		status := StatusUnknown
		if envelope.Verdict == VerdictFailClosed || envelope.Verifier.State == VerifierFailClosed {
			status = StatusFailClosed
		}
		return Explanation{Status: status, EvidenceDigest: envelope.EvidenceDigest, Binding: binding,
			Upstream: envelope.Upstream, NoLink: &NoLink{Reason: envelope.NoLinkReason}, Diagnostics: canonicalDiagnostics(envelope.Diagnostics)}
	}
	link := ExplanationLink{CodeBinding: envelope.CodeBinding, SemanticOwner: envelope.SemanticOwner,
		Term: envelope.Term, OriginPath: envelope.OriginPath, Receipt: envelope.Receipt, Verifier: envelope.Verifier}
	return Explanation{Status: StatusPass, EvidenceDigest: envelope.EvidenceDigest, Binding: binding,
		Upstream: envelope.Upstream, Link: &link}
}
func requestBinding(request Request) SnapshotBinding {
	return SnapshotBinding{SnapshotDigest: request.SnapshotDigest, RegistryDigest: request.RegistryDigest,
		SourceMapDigest: request.SourceMapDigest, ManifestDigest: request.ManifestDigest, ToolchainDigest: request.ToolchainDigest,
		ProfileDigest: request.ProfileDigest, DetectorInputDigest: request.DetectorInputDigest, DetectorResultDigest: request.DetectorResultDigest,
		VerifierResultDigest: request.VerifierResultDigest, EnvelopeDigest: request.EnvelopeDigest, Control: request.Control}
}
