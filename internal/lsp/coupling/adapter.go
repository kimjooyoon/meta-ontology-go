package coupling

import (
	"context"
	"sort"
)

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

func (a *Adapter) Hover(request Request) (*Hover, []Diagnostic) {
	result := a.Resolve(request)
	return result.Hover, result.Diagnostics
}

func (a *Adapter) Diagnostics(request Request) []Diagnostic {
	return a.Resolve(request).Diagnostics
}

func (a *Adapter) matches(position Position) []Explanation {
	matches := make([]Explanation, 0, 1)
	for _, explanation := range a.envelope.Explanations {
		if positionInRange(position, explanation.Origin.Range) {
			matches = append(matches, explanation)
		}
	}
	return matches
}

func success(explanation Explanation) Result {
	origin := explanation.Origin.Range
	links := []LocationLink{{
		OriginSelectionRange: &origin,
		TargetURI:            explanation.Target.URI,
		TargetRange:          explanation.Target.Range,
		TargetSelectionRange: explanation.Target.Range,
	}}

	title := explanation.Label
	if title == "" {
		title = "Semantic coupling explanation"
	}
	hover := &Hover{
		Contents: MarkupContent{Kind: "plaintext", Value: title + "\nClaim: " + string(explanation.Claim)},
		Range:    &origin,
	}

	spans := append([]CausalSpan(nil), explanation.CausalSpans...)
	sort.SliceStable(spans, func(left, right int) bool {
		if spans[left].Ordinal != spans[right].Ordinal {
			return spans[left].Ordinal < spans[right].Ordinal
		}
		return spans[left].StableID < spans[right].StableID
	})
	related := make([]DiagnosticRelatedInformation, 0, len(spans))
	for _, span := range spans {
		message := span.Message
		if message == "" {
			message = "Contributing causal span."
		}
		related = append(related, DiagnosticRelatedInformation{
			Location: Location{URI: span.URI, Range: span.Range}, Message: message,
		})
	}
	diagnostic := Diagnostic{
		Range: origin, Severity: DiagnosticInformation, Code: DiagnosticExplanation,
		Source: diagnosticSource, Message: "Coupling explanation is current.",
		RelatedInformation: related,
	}
	return Result{Outcome: OutcomePass, Links: links, Hover: hover, Diagnostics: []Diagnostic{diagnostic}}
}

func failureForReason(outcome Outcome, reason Reason, span Range) Result {
	code, severity, message := reasonDiagnostic(reason)
	if outcome == OutcomeFailClosed {
		severity = DiagnosticError
	}
	return failure(outcome, code, severity, message, span)
}

func reasonDiagnostic(reason Reason) (string, int, string) {
	switch reason {
	case ReasonAmbiguous:
		return DiagnosticAmbiguous, DiagnosticWarning, "The coupling explanation is ambiguous."
	case ReasonStaleSnapshot:
		return DiagnosticStaleSnapshot, DiagnosticWarning, "The coupling explanation is stale."
	case ReasonUpstreamUnknown:
		return DiagnosticUpstreamUnknown, DiagnosticWarning, "The upstream coupling explanation is unknown."
	case ReasonUpstreamFail:
		return DiagnosticUpstreamFail, DiagnosticError, "The upstream coupling explanation failed closed."
	case ReasonUnregistered:
		return DiagnosticNoBinding, DiagnosticWarning, "The code surface is not registered in the coupling snapshot."
	case ReasonMissing:
		return DiagnosticNoBinding, DiagnosticWarning, "The coupling explanation is missing."
	default:
		return DiagnosticInvalidEnvelope, DiagnosticError, "The coupling explanation is invalid."
	}
}

func failure(outcome Outcome, code string, severity int, message string, span Range) Result {
	return Result{
		Outcome: outcome,
		Diagnostics: []Diagnostic{{
			Range: span, Severity: severity, Code: code, Source: diagnosticSource, Message: message,
		}},
	}
}

func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func positionInRange(position Position, value Range) bool {
	if position.Line < value.Start.Line || position.Line > value.End.Line {
		return false
	}
	if position.Line == value.Start.Line && position.Character < value.Start.Character {
		return false
	}
	if position.Line == value.End.Line && position.Character >= value.End.Character {
		return false
	}
	return true
}
