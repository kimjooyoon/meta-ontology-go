package coupling

// Resolve adapts one exact source position to standard LSP navigation,
// hover, and diagnostic values. It performs no name lookup and never reads a
// file or writes a workspace. Cancellation is checked before every response
// phase; cancellation wins over a potentially valid explanation.
func (a *Adapter) Resolve(request Request) Result {
	if a == nil {
		return failure(OutcomeFailClosed, DiagnosticInvalidEnvelope, DiagnosticError, "Coupling explanation is unavailable.", Range{})
	}
	if request.Context == nil {
		return failure(OutcomeUnknown, DiagnosticMissingCancellation, DiagnosticWarning, "A cancellation token is required.", Range{})
	}
	if cancelled(request.Context) {
		return failure(OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling explanation request was cancelled.", Range{})
	}
	if request.DocumentURI == "" {
		return failure(OutcomeUnknown, DiagnosticDocumentMismatch, DiagnosticWarning, "The document URI is required.", Range{})
	}
	if request.DocumentVersion <= 0 {
		return failure(OutcomeUnknown, DiagnosticMissingVersion, DiagnosticWarning, "The exact document version is required.", Range{})
	}
	if request.SnapshotDigest == "" {
		return failure(OutcomeUnknown, DiagnosticMissingSnapshot, DiagnosticWarning, "The exact snapshot digest is required.", Range{})
	}
	if request.Position.Line < 0 || request.Position.Character < 0 {
		return failure(OutcomeUnknown, DiagnosticInvalidPosition, DiagnosticWarning, "The document position is invalid.", Range{})
	}
	if request.DocumentURI != a.envelope.Document.URI {
		return failure(OutcomeUnknown, DiagnosticDocumentMismatch, DiagnosticWarning, "The requested document is not the explanation document.", Range{})
	}
	if request.DocumentVersion != a.envelope.Document.Version {
		return failure(OutcomeUnknown, DiagnosticWrongVersion, DiagnosticWarning, "The document version is stale.", Range{})
	}
	if request.SnapshotDigest != a.envelope.SnapshotDigest {
		return failure(OutcomeUnknown, DiagnosticStaleSnapshot, DiagnosticWarning, "The explanation snapshot is stale.", Range{})
	}
	if cancelled(request.Context) {
		return failure(OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling explanation request was cancelled.", Range{})
	}
	if a.envelope.Status != OutcomePass {
		return failureForReason(a.envelope.Status, a.envelope.Reason, Range{})
	}

	matches := a.matches(request.Position)
	if len(matches) == 0 {
		return failure(OutcomeUnknown, DiagnosticNoBinding, DiagnosticWarning, "No exact coupling binding covers this position.", Range{})
	}
	if len(matches) != 1 {
		return failure(OutcomeUnknown, DiagnosticAmbiguous, DiagnosticWarning, "The coupling binding is ambiguous.", Range{})
	}
	if cancelled(request.Context) {
		return failure(OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling explanation request was cancelled.", matches[0].Origin.Range)
	}

	explanation := matches[0]
	if explanation.Status != OutcomePass {
		return failureForReason(explanation.Status, explanation.Reason, explanation.Origin.Range)
	}
	return success(explanation)
}

// Explain is a descriptive alias for Resolve for callers that treat the LSP
// adapter as a coupling explanation view rather than a server feature.
func (a *Adapter) Explain(request Request) Result { return a.Resolve(request) }

// Definition returns only standard LocationLink values. Any UNKNOWN,
// FAIL_CLOSED, stale, ambiguous, or cancelled result returns no links and a
// deterministic diagnostic.
func (a *Adapter) Definition(request Request) ([]LocationLink, []Diagnostic) {
	result := a.Resolve(request)
	return result.Links, result.Diagnostics
}
