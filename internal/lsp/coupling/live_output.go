package coupling

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
)

func liveSuccess(request LiveRequest, explanation couplingexplain.Explanation) Result {
	link := *explanation.Link
	locations := make(map[string]SourceLocation, len(request.Locations.Locations))
	for _, location := range request.Locations.Locations {
		locations[location.StableID] = location
	}
	origin := locations[link.CodeBinding.CodeSymbolID]
	target := locations[link.Term.TermID]
	originRange := origin.Range
	links := []LocationLink{{OriginSelectionRange: &originRange, TargetURI: target.URI, TargetRange: target.Range, TargetSelectionRange: target.Range}}
	related := liveRelatedInformation(link, locations)
	message := "Coupling explanation is current."
	if link.Receipt.ChangeClaim == couplingexplain.ClaimDelta {
		message = "Coupling explanation reports a semantic delta."
	} else if link.Receipt.ChangeClaim == couplingexplain.ClaimNoDelta {
		message = "Coupling explanation reports no semantic delta."
	}
	return Result{Outcome: OutcomePass, Links: links,
		Hover: &Hover{Contents: MarkupContent{Kind: "plaintext", Value: liveHoverText(link)}, Range: &originRange},
		Diagnostics: []Diagnostic{{Range: originRange, Severity: DiagnosticInformation,
			Code: DiagnosticExplanation, Source: diagnosticSource, Message: message, RelatedInformation: related}}}
}

func liveHoverText(link couplingexplain.ExplanationLink) string {
	label := link.Term.Presentation.Label
	if label == "" {
		label = link.CodeBinding.Presentation.Label
	}
	if label == "" {
		label = "Semantic coupling explanation"
	}
	return label + "\nClaim: " + string(link.Receipt.ChangeClaim)
}

func liveRelatedInformation(link couplingexplain.ExplanationLink, locations map[string]SourceLocation) []DiagnosticRelatedInformation {
	ids := make([]string, 0, len(link.OriginPath.Steps)*2+len(link.Receipt.EvidenceRefs))
	seen := make(map[string]struct{})
	add := func(id string) {
		if id == "" {
			return
		}
		if _, exists := seen[id]; exists {
			return
		}
		if _, ok := locations[id]; !ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	for _, step := range link.OriginPath.Steps {
		add(step.ToID)
		add(step.EvidenceRef)
	}
	for _, id := range link.Receipt.EvidenceRefs {
		add(id)
	}
	sort.Strings(ids)
	related := make([]DiagnosticRelatedInformation, 0, len(ids))
	for _, id := range ids {
		location := locations[id]
		message := location.Message
		if message == "" {
			message = "Contributing verified coupling evidence."
		}
		related = append(related, DiagnosticRelatedInformation{Location: Location{URI: location.URI, Range: location.Range}, Message: message})
	}
	return related
}

func liveFailure(request LiveRequest, outcome Outcome, code string, severity int, message string) Result {
	position := request.Position
	if position.Line < 0 || position.Character < 0 {
		position = Position{}
	}
	span := Range{Start: position, End: position}
	return Result{Outcome: outcome, Diagnostics: []Diagnostic{{Range: span, Severity: severity, Code: code, Source: diagnosticSource, Message: message}}}
}
