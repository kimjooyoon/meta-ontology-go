package coupling

import (
	"context"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
)

// SourceLocation is a source-map-backed location supplied by the caller. It
// is not inferred from a query path, name, alias, or presentation field.
// StableID and SourceMapID are validation inputs and never appear on the LSP
// wire.
type SourceLocation struct {
	StableID        string
	SourceMapID     string
	SourceMapDigest string
	URI             string
	Range           Range
	Label           string
	Message         string
}

// LocationSnapshot binds every source location used by the adapter to one
// exact snapshot and LSP document version. The production query envelope
// does not contain URI/range data, so a verified link cannot be emitted
// without this explicit projection.
type LocationSnapshot struct {
	SnapshotDigest  string
	DocumentURI     string
	DocumentVersion int
	Locations       []SourceLocation
}

// LiveRequest combines the LSP request with the immutable query request. The
// query request's control versions must be equal and must match the exact LSP
// document version before a link can be returned.
type LiveRequest struct {
	Context         context.Context
	DocumentURI     string
	DocumentVersion int
	Position        Position
	SnapshotDigest  string
	Query           couplingexplain.Request
	Locations       LocationSnapshot
}

// LiveAdapter consumes only the production couplingexplain envelope bytes.
// It has no write, rename, filesystem, graph, or semantic-authority surface.
type LiveAdapter struct {
	raw []byte
}

// NewLiveAdapter validates and copies one immutable production query envelope.
func NewLiveAdapter(data []byte) (*LiveAdapter, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("coupling query: empty envelope")
	}
	if _, err := couplingexplain.DecodeVerifiedEnvelope(data); err != nil {
		return nil, fmt.Errorf("coupling query: decode: %w", err)
	}
	return &LiveAdapter{raw: append([]byte(nil), data...)}, nil
}

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

func validateLiveLocations(request LiveRequest, link couplingexplain.ExplanationLink) *liveIssue {
	if len(request.Locations.Locations) == 0 {
		return &liveIssue{OutcomeUnknown, DiagnosticLiveMissingLocations, DiagnosticWarning, "The verified query has no source locations for LSP navigation."}
	}
	byID := make(map[string]SourceLocation, len(request.Locations.Locations))
	for _, location := range request.Locations.Locations {
		if err := validateSourceLocation(request, location); err != nil {
			return err
		}
		if _, exists := byID[location.StableID]; exists {
			return &liveIssue{OutcomeUnknown, DiagnosticAmbiguous, DiagnosticWarning, "The source location binding is ambiguous."}
		}
		byID[location.StableID] = location
	}
	origin, originOK := byID[link.CodeBinding.CodeSymbolID]
	target, targetOK := byID[link.Term.TermID]
	if !originOK || !targetOK {
		return &liveIssue{OutcomeUnknown, DiagnosticLiveMissingLocations, DiagnosticWarning, "The verified query lacks an exact origin or target source location."}
	}
	if origin.SourceMapID != link.CodeBinding.SourceMapID || origin.URI != request.DocumentURI {
		return &liveIssue{OutcomeUnknown, DiagnosticStaleSnapshot, DiagnosticWarning, "The origin location is not bound to the verified query."}
	}
	if origin.StableID != link.CodeBinding.CodeSymbolID || target.StableID != link.Term.TermID {
		return &liveIssue{OutcomeUnknown, DiagnosticAmbiguous, DiagnosticWarning, "The verified query location binding is ambiguous."}
	}
	for _, stableID := range liveRequiredLocationIDs(link) {
		if _, ok := byID[stableID]; !ok {
			return &liveIssue{OutcomeUnknown, DiagnosticLiveMissingLocations, DiagnosticWarning, "The verified query lacks a required contributing source location."}
		}
	}
	return nil
}

func liveRequiredLocationIDs(link couplingexplain.ExplanationLink) []string {
	values := []string{link.CodeBinding.CodeSymbolID, link.SemanticOwner, link.Term.TermID,
		link.OriginPath.StartID, link.OriginPath.EndID, link.Verifier.EvidenceID}
	for _, step := range link.OriginPath.Steps {
		values = append(values, step.ToID, step.EvidenceRef)
	}
	values = append(values, link.Receipt.EvidenceRefs...)
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func validateSourceLocation(request LiveRequest, location SourceLocation) *liveIssue {
	if err := validateIdentity(location.StableID, "source location"); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location stable ID is malformed."}
	}
	if err := validateIdentity(location.SourceMapID, "source map"); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location map ID is malformed."}
	}
	if err := validateDigest(location.SourceMapDigest); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location map digest is malformed."}
	}
	if location.SourceMapDigest != request.Query.SourceMapDigest {
		return &liveIssue{OutcomeUnknown, DiagnosticStaleSnapshot, DiagnosticWarning, "The source location map is stale."}
	}
	if location.URI == "" || strings.TrimSpace(location.URI) != location.URI {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location URI is malformed."}
	}
	if err := validateRange(location.Range); err != nil {
		return &liveIssue{OutcomeFailClosed, DiagnosticLiveInvalidBinding, DiagnosticError, "The source location range is malformed."}
	}
	return nil
}

func liveQueryFailure(request LiveRequest, explanation couplingexplain.Explanation) Result {
	outcome := OutcomeUnknown
	if explanation.Status == couplingexplain.StatusFailClosed {
		outcome = OutcomeFailClosed
	}
	code, severity, message := liveReasonDiagnostic(explanation.NoLink)
	if outcome == OutcomeFailClosed {
		severity = DiagnosticError
	}
	return liveFailure(request, outcome, code, severity, message)
}

func liveReasonDiagnostic(noLink *couplingexplain.NoLink) (string, int, string) {
	if noLink == nil {
		return DiagnosticLiveInvalidEnvelope, DiagnosticError, "The coupling query did not return a link decision."
	}
	switch noLink.Reason {
	case couplingexplain.ReasonAmbiguous:
		return DiagnosticAmbiguous, DiagnosticWarning, "The coupling query is ambiguous."
	case couplingexplain.ReasonStale:
		return DiagnosticStaleSnapshot, DiagnosticWarning, "The coupling query is stale."
	case couplingexplain.ReasonUnregistered:
		return DiagnosticNoBinding, DiagnosticWarning, "The code surface is not registered in the coupling snapshot."
	case couplingexplain.ReasonMissing:
		return DiagnosticLiveMissingMaterial, DiagnosticWarning, "The coupling query lacks verified link material."
	case couplingexplain.ReasonNotVerified:
		return DiagnosticLiveNotVerified, DiagnosticError, "The coupling query link was not independently verified."
	default:
		return DiagnosticLiveInvalidEnvelope, DiagnosticError, "The coupling query returned an invalid link reason."
	}
}
