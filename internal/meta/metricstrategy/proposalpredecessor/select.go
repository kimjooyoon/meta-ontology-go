package proposalpredecessor

import "fmt"

func Select(repository, currentSHA, predecessorSHA string, collection Collection) (Report, []byte, error) {
	summary := Summary{
		ObservedRuns: collection.ObservedRuns, ExactRuns: collection.ExactRuns,
		OtherRouteRuns: collection.OtherRouteRuns, RouteUnknownRuns: collection.RouteUnknownRuns,
		Contradictions: collection.Contradictions, ObservedArtifacts: collection.ObservedArtifacts,
		ExactArtifacts: collection.ExactArtifacts, ObservedJobs: collection.ObservedJobs,
		ExactJobs: collection.ExactJobs, ValidCandidates: len(collection.Candidates),
		UnresolvedCandidates: collection.Unresolved, ProofsTotal: 5,
	}
	if len(collection.Candidates) > 1 {
		summary.AmbiguousCandidates = len(collection.Candidates) - 1
	}
	if len(collection.Candidates) == 0 && summary.UnresolvedCandidates == 0 {
		summary.UnresolvedCandidates = 1
	}
	decision, reason := "FAIL_CLOSED", ReasonNotFound
	observationDecision, observationResolution := DecisionUnknown, ResolutionLower
	var unknown *Unknown
	var selected *Selected
	var payload []byte
	if collection.Contradictions > 0 {
		reason = ReasonRouteContradiction
		observationDecision, observationResolution = DecisionRefuted, ResolutionExact
	} else if collection.RouteUnknownRuns > 0 {
		reason = ReasonRouteUnknown
		unknown = routeUnknown(reason, collection.RequestedRoute)
	} else if len(collection.Candidates) > 1 {
		reason = ReasonAmbiguous
		unknown = routeAmbiguous(reason, collection.RequestedRoute)
	} else if len(collection.Candidates) == 1 && summary.UnresolvedCandidates != 0 {
		reason = ReasonEvidenceUnknown
		unknown = routeEvidenceUnknown(reason, collection.RequestedRoute)
	} else if len(collection.Candidates) == 1 && summary.ExactJobs != 1 {
		reason = ReasonJobCardinality
		unknown = routeEvidenceUnknown(reason, collection.RequestedRoute)
	} else if len(collection.Candidates) == 1 && candidateReady(collection.Candidates[0], predecessorSHA) {
		candidate := collection.Candidates[0]
		selected, payload = &candidate.Selected, candidate.ProposalPayload
		decision, reason, summary.SelectionBPS = "SELECTED", ReasonSelected, 10000
		observationDecision, observationResolution = DecisionClosed, ResolutionExact
		unknown = nil
	} else if collection.FailureReason != "" {
		reason = collection.FailureReason
		if reason == ReasonRouteContradiction {
			observationDecision, observationResolution = DecisionRefuted, ResolutionExact
		} else {
			unknown = routeEvidenceUnknown(reason, collection.RequestedRoute)
		}
	} else {
		unknown = routeMissing(reason, collection.RequestedRoute)
	}
	proofs, err := buildProofs(selected, predecessorSHA, summary)
	if err != nil {
		return Report{}, nil, err
	}
	for _, proof := range proofs {
		if proof.Passed {
			summary.ProofsPassed++
		}
	}
	report := Report{
		Schema: Schema, Repository: repository, CurrentSubjectSHA: currentSHA,
		PredecessorSHA: predecessorSHA, RequestedRoute: collection.RequestedRoute,
		Decision: decision, Reason: reason, ObservationDecision: observationDecision,
		ObservationResolution: observationResolution, Unknown: unknown, Selected: selected,
		Summary: summary, Indicators: buildIndicators(summary), Proofs: proofs,
	}
	report, err = sealReport(report)
	if err != nil {
		return Report{}, nil, err
	}
	if err := Validate(report); err != nil {
		return Report{}, nil, err
	}
	if !report.Ready() {
		return report, nil, fmt.Errorf("proposal predecessor selection failed closed: %s", report.Reason)
	}
	return report, payload, nil
}

func routeUnknown(reason, route string) *Unknown {
	return &Unknown{Stage: ResolutionStage, Step: ResolutionStep, Reason: reason,
		UnknownClass: UnknownClassRoute, NextOperation: "obtain route-bound run identity",
		BlockedBy: []string{"head_branch"}}
}

func routeAmbiguous(reason, route string) *Unknown {
	return &Unknown{Stage: ResolutionStage, Step: ResolutionStep, Reason: reason,
		UnknownClass: UnknownClassAmbiguous, NextOperation: "resolve one exact requested-route candidate",
		BlockedBy: []string{"requested_route:" + route, "candidate_count>1"}}
}

func routeEvidenceUnknown(reason, route string) *Unknown {
	return &Unknown{Stage: ResolutionStage, Step: ResolutionStep, Reason: reason,
		UnknownClass: UnknownClassRoute, NextOperation: "complete route-bound predecessor observation",
		BlockedBy: []string{"requested_route:" + route}}
}

func routeMissing(reason, route string) *Unknown {
	return &Unknown{Stage: ResolutionStage, Step: ResolutionStep, Reason: reason,
		UnknownClass: UnknownClassMissing, NextOperation: "obtain one route-bound predecessor artifact",
		BlockedBy: []string{"requested_route:" + route}}
}

func candidateReady(candidate Candidate, predecessorSHA string) bool {
	selected := candidate.Selected
	return selected.HeadSHA == predecessorSHA && validRoute(selected.HeadBranch) && selected.ContractSatisfied == 8 && selected.ContractTotal == 8 && selected.ContractBPS == 10000 && selected.ContractUnresolved == 0 && selected.RepositoryWrites == 0 && !selected.PromotionAuthorized && canonicalSelected(selected)
}
