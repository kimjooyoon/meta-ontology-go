package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
)

// ResolveLive is the safe one-shot entry point. Invalid or malformed input
// becomes a deterministic fail-closed diagnostic and never panics.
func ResolveLive(request LiveRequest, data []byte) Result {
	if request.Context == nil {
		return liveFailure(request, OutcomeUnknown, DiagnosticMissingCancellation, DiagnosticWarning, "A cancellation token is required.")
	}
	if cancelled(request.Context) {
		return liveFailure(request, OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling query request was cancelled.")
	}
	if issue := validateLiveRequest(request); issue != nil {
		return liveFailure(request, issue.outcome, issue.code, issue.severity, issue.message)
	}
	adapter, err := NewLiveAdapter(data)
	if err != nil {
		return liveFailure(request, OutcomeFailClosed, DiagnosticLiveInvalidEnvelope, DiagnosticError, "The coupling query envelope is invalid.")
	}
	return adapter.Resolve(request)
}

// RawBytes returns a copy of the immutable query bytes.
func (a *LiveAdapter) RawBytes() []byte {
	if a == nil {
		return nil
	}
	return append([]byte(nil), a.raw...)
}

// Resolve projects one live query explanation to standard LSP values. A
// query PASS is insufficient by itself: exact source-map locations, snapshot,
// URI, and version binding are also required before a link is emitted.
func (a *LiveAdapter) Resolve(request LiveRequest) Result {
	if a == nil {
		return liveFailure(request, OutcomeFailClosed, DiagnosticLiveInvalidEnvelope, DiagnosticError, "The coupling query adapter is unavailable.")
	}
	if request.Context == nil {
		return liveFailure(request, OutcomeUnknown, DiagnosticMissingCancellation, DiagnosticWarning, "A cancellation token is required.")
	}
	if cancelled(request.Context) {
		return liveFailure(request, OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling query request was cancelled.")
	}
	if issue := validateLiveRequest(request); issue != nil {
		return liveFailure(request, issue.outcome, issue.code, issue.severity, issue.message)
	}

	explanation, err := couplingexplain.ExplainEnvelopeBytes(request.Context, request.Query, a.raw)
	if err != nil {
		return liveFailure(request, OutcomeFailClosed, DiagnosticLiveInvalidEnvelope, DiagnosticError, "The coupling query envelope could not be decoded.")
	}
	if cancelled(request.Context) {
		return liveFailure(request, OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling query request was cancelled.")
	}
	if explanation.Status != couplingexplain.StatusPass || explanation.Link == nil {
		return liveQueryFailure(request, explanation)
	}
	if issue := validateLiveLocations(request, *explanation.Link); issue != nil {
		return liveFailure(request, issue.outcome, issue.code, issue.severity, issue.message)
	}
	if cancelled(request.Context) {
		return liveFailure(request, OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling query request was cancelled.")
	}
	return liveSuccess(request, explanation)
}

// Definition returns only standard LocationLink values and diagnostics.
func (a *LiveAdapter) Definition(request LiveRequest) ([]LocationLink, []Diagnostic) {
	result := a.Resolve(request)
	return result.Links, result.Diagnostics
}
