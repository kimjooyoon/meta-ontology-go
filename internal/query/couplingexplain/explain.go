package couplingexplain

import (
	"context"
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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
	binding := SnapshotBinding{SnapshotDigest: request.SnapshotDigest, RegistryDigest: request.RegistryDigest,
		SourceMapDigest: request.SourceMapDigest, ManifestDigest: request.ManifestDigest, ToolchainDigest: request.ToolchainDigest, ProfileDigest: request.ProfileDigest,
		DetectorInputDigest: request.DetectorInputDigest, DetectorResultDigest: request.DetectorResultDigest, VerifierResultDigest: request.VerifierResultDigest,
		EnvelopeDigest: request.EnvelopeDigest, Control: request.Control}
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

func validateRequestBinding(request Request, envelope VerifiedEnvelope) *envelopeIssue {
	if request.CodeSymbolID == "" || !validDigest(request.SnapshotDigest) || !validDigest(request.RegistryDigest) ||
		!validDigest(request.SourceMapDigest) || !validDigest(request.ManifestDigest) || !validDigest(request.ToolchainDigest) || !validDigest(request.ProfileDigest) ||
		!validDigest(request.DetectorInputDigest) || !validDigest(request.DetectorResultDigest) || !validDigest(request.VerifierResultDigest) || !validDigest(request.EnvelopeDigest) {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonStale, code: "invalid-request-binding"}
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

func validateEnvelopeShape(envelope VerifiedEnvelope) *envelopeIssue {
	if envelope.Schema == "" || envelope.Binding.SnapshotDigest == "" || !validDigest(envelope.EvidenceDigest) {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonMissing, code: "incomplete-envelope-header"}
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
	if envelope.CodeBinding.CodeSymbolID == "" || envelope.CodeBinding.SemanticOwnerID == "" || envelope.SemanticOwner == "" ||
		envelope.Term.TermID == "" || envelope.Term.Version == "" ||
		!validDigest(envelope.Term.DefinitionDigest) || envelope.OriginPath.PathID == "" ||
		envelope.OriginPath.StepCount != len(envelope.OriginPath.Steps) || envelope.OriginPath.StepCount < 1 ||
		!validDigest(envelope.OriginPath.PathDigest) || envelope.Receipt.ReceiptID == "" ||
		envelope.Receipt.SurfaceID == "" || !validChangeClaim(envelope.Receipt.ChangeClaim) ||
		!envelope.Receipt.ReceiptKind.Valid() || envelope.Receipt.OriginPathID == "" ||
		envelope.Verifier.EvidenceID == "" || envelope.Verifier.ReceiptID == "" ||
		!validDigest(envelope.Verifier.EvidenceDigest) {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonMissing, code: "incomplete-verified-envelope"}
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
	return nil
}

func validateConnectedPath(path PathSummary, ownerID, termID string) *envelopeIssue {
	visited := map[string]struct{}{path.StartID: {}}
	for index, step := range path.Steps {
		if step.FromID == "" || step.ToID == "" || step.FromID == step.ToID || !step.Kind.Valid() ||
			!step.Phase.Phase.Valid() || !validDigest(step.InputDigest) || !validDigest(step.OutputDigest) {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "malformed-origin-path", ids: []string{path.PathID}}
		}
		if index == 0 && step.FromID != path.StartID {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "disconnected-origin-path", ids: []string{path.PathID}}
		}
		if index > 0 && step.FromID != path.Steps[index-1].ToID {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "disconnected-origin-path", ids: []string{path.PathID}}
		}
		if _, exists := visited[step.ToID]; exists {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "cyclic-origin-path", ids: []string{path.PathID}}
		}
		visited[step.ToID] = struct{}{}
		if step.Kind == semantic.InferenceObservationCandidate {
			return &envelopeIssue{status: StatusFailClosed, reason: ReasonAmbiguous, code: "candidate-in-verified-path", ids: []string{path.PathID}}
		}
	}
	last := path.Steps[len(path.Steps)-1]
	if path.Steps[0].ToID != ownerID {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonUnregistered, code: "origin-owner-mismatch", ids: []string{path.PathID}}
	}
	termReached := false
	for _, step := range path.Steps {
		if step.FromID == ownerID && step.ToID == termID {
			termReached = true
		}
	}
	if !termReached {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonMissing, code: "origin-term-missing", ids: []string{path.PathID}}
	}
	if last.ToID != path.EndID || last.Kind != semantic.InferenceIndependentVerification || last.EvidenceRef == "" {
		return &envelopeIssue{status: StatusFailClosed, reason: ReasonNotVerified, code: "origin-path-verifier-missing", ids: []string{path.PathID}}
	}
	return nil
}

func requestNoLink(binding SnapshotBinding, status Status, reason LinkReason, code string, ids ...string) Explanation {
	diagnostics := []Diagnostic{{Code: code, IDs: sortedStrings(ids)}}
	value := struct {
		Status      Status          `json:"status"`
		Binding     SnapshotBinding `json:"binding"`
		NoLink      NoLink          `json:"no_link"`
		Diagnostics []Diagnostic    `json:"diagnostics"`
	}{Status: status, Binding: binding, NoLink: NoLink{Reason: reason}, Diagnostics: diagnostics}
	data, _ := json.Marshal(value)
	return Explanation{Status: status, EvidenceDigest: DigestBytes(data), Binding: binding,
		NoLink: &NoLink{Reason: reason}, Diagnostics: diagnostics}
}

func validChangeClaim(value ChangeClaim) bool { return value == ClaimDelta || value == ClaimNoDelta }

func validReason(value LinkReason) bool {
	switch value {
	case ReasonAmbiguous, ReasonStale, ReasonUnregistered, ReasonMissing, ReasonNotVerified:
		return true
	default:
		return false
	}
}
