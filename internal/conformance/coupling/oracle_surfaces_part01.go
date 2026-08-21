package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func inputObservationCounts(input Input) ObservationCounts {
	counts := ObservationCounts{
		RegistryBindings: uint64(len(input.Registry)), ReceiptRecords: uint64(len(input.Receipts)),
		PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence)),
	}
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			counts.ChangedCodeSurfaces++
		}
	}
	for _, edge := range input.Path.Edges {
		switch edge.Kind {
		case semantic.InferenceObservationCandidate:
			counts.CandidateObservations++
		case semantic.InferenceAcceptedLift:
			counts.AcceptedLifts++
		}
	}
	return counts
}
func normalizeRegistry(bindings []CodeBinding) (registryView, oracleValidation) {
	view := registryView{bySurface: make(map[string]CodeBinding), bySymbol: make(map[string]CodeBinding)}
	canonical := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if !validID(binding.RegisteredSurfaceID) || !validID(binding.CodeSymbolID) || !validID(binding.SemanticOwnerID) ||
			!validToken(binding.SourceMapID) || !validDigest(binding.BindingDigest) {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if _, exists := view.bySurface[binding.RegisteredSurfaceID]; exists {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if _, exists := view.bySymbol[binding.CodeSymbolID]; exists {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		if expected := bindingDigest(binding); expected != binding.BindingDigest {
			return registryView{}, oracleValidation{DecisionFailClosed, ReasonRegistryBinding}
		}
		view.bySurface[binding.RegisteredSurfaceID] = binding
		view.bySymbol[binding.CodeSymbolID] = binding
		canonical = append(canonical, bindingCanonical(binding))
	}
	sort.Strings(canonical)
	view.digest = digestBytes([]byte(strings.Join(canonical, "\n") + "\n"))
	return view, oracleValidation{}
}
