package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validatePath(input Input, registry registryView, receipts map[string]CouplingReceipt, beforeDigest, afterDigest, deltaText string) pathView {
	view := pathView{counts: ObservationCounts{PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence))}}
	if len(receipts) == 0 {
		if len(input.Path.Edges) == 0 && len(input.Path.Claims) == 0 && len(input.Path.Evidence) == 0 && len(input.Roots) == 0 {
			view.digest = pathDigest(input.Path)
			return view
		}
		view.decision, view.reason = DecisionFailClosed, ReasonPathClosure
		return view
	}
	if input.Path.Version != semantic.InferencePathSchemaVersion {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	roots, issue := parseUniqueIDs(input.Roots)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	if len(roots) == 0 {
		view.decision, view.reason = DecisionUnknown, ReasonPathMissing
		return view
	}
	rootSet := make(map[semantic.ID]struct{}, len(roots))
	for _, root := range roots {
		rootSet[root] = struct{}{}
	}
	evidence, issue := collectEvidence(input.Path.Evidence, beforeDigest, afterDigest, input.Config)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	edges, issue := collectEdges(input.Path.Edges, evidence, beforeDigest, afterDigest, input.Config)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	claims, issue := collectClaims(input.Path.Claims, evidence, beforeDigest, afterDigest, deltaText)
	if issue != "" {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMalformed
		return view
	}
	if !declarationRootsMatch(edges, rootSet) {
		view.decision, view.reason = DecisionFailClosed, ReasonPathMissing
		return view
	}
	if !receiptsClosePath(roots, receipts, registry, edges, claims, evidence) {
		view.decision, view.reason = DecisionFailClosed, ReasonPathClosure
		return view
	}
	view.digest = pathDigest(input.Path)
	for _, edge := range input.Path.Edges {
		switch edge.Kind {
		case semantic.InferenceObservationCandidate:
			view.counts.CandidateObservations++
		case semantic.InferenceAcceptedLift:
			view.counts.AcceptedLifts++
		}
	}
	return view
}
