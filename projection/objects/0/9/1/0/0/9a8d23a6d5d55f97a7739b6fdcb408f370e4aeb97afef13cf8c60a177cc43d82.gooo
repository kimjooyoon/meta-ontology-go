package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
)

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
