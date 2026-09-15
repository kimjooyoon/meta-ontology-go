package coupling

import (
	"strings"
)

// Hover returns only standard Hover values and diagnostics.
func (a *LiveAdapter) Hover(request LiveRequest) (*Hover, []Diagnostic) {
	result := a.Resolve(request)
	return result.Hover, result.Diagnostics
}

// Diagnostics returns only standard Diagnostic values.
func (a *LiveAdapter) Diagnostics(request LiveRequest) []Diagnostic {
	return a.Resolve(request).Diagnostics
}

type liveIssue struct {
	outcome  Outcome
	code     string
	severity int
	message  string
}

func validateLiveRequest(request LiveRequest) *liveIssue {
	if request.Query.Control.RequestCancellationVersion != request.Query.Control.ObservedCancellationVersion {
		return &liveIssue{OutcomeUnknown, DiagnosticCancelled, DiagnosticWarning, "The coupling query cancellation ordering is stale."}
	}
	if request.Query.Control.RequestVersion != request.Query.Control.ObservedVersion {
		return &liveIssue{OutcomeUnknown, DiagnosticWrongVersion, DiagnosticWarning, "The coupling query version ordering is stale."}
	}
	if request.DocumentURI == "" || strings.TrimSpace(request.DocumentURI) != request.DocumentURI {
		return &liveIssue{OutcomeUnknown, DiagnosticDocumentMismatch, DiagnosticWarning, "The exact document URI is required."}
	}
	if request.DocumentVersion <= 0 || uint64(request.DocumentVersion) != request.Query.Control.RequestVersion {
		return &liveIssue{OutcomeUnknown, DiagnosticWrongVersion, DiagnosticWarning, "The exact document version is stale or missing."}
	}
	if request.Position.Line < 0 || request.Position.Character < 0 {
		return &liveIssue{OutcomeUnknown, DiagnosticInvalidPosition, DiagnosticWarning, "The document position is invalid."}
	}
	if request.SnapshotDigest == "" || request.SnapshotDigest != request.Query.SnapshotDigest {
		return &liveIssue{OutcomeUnknown, DiagnosticStaleSnapshot, DiagnosticWarning, "The coupling query snapshot is stale."}
	}
	if request.Locations.DocumentURI != request.DocumentURI || request.Locations.DocumentVersion != request.DocumentVersion {
		return &liveIssue{OutcomeUnknown, DiagnosticStaleSnapshot, DiagnosticWarning, "The source locations are bound to a different document version."}
	}
	if request.Locations.SnapshotDigest == "" || request.Locations.SnapshotDigest != request.SnapshotDigest {
		return &liveIssue{OutcomeUnknown, DiagnosticStaleSnapshot, DiagnosticWarning, "The source locations are bound to a different snapshot."}
	}
	if err := validateDigest(request.SnapshotDigest); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The coupling query snapshot binding is malformed."}
	}
	return nil
}
