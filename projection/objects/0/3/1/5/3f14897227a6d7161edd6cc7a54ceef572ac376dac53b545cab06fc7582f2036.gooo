package coupling

import (
	"sort"
)

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
