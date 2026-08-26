package externalconformanceactivation

import "encoding/json"

func validateEvidence(input Input) validation {
	bindings := artifactBindings(input)
	state := validation{Resolution: ResolutionUnknown}
	for _, binding := range bindings {
		if binding.Exact {
			state.RawExact++
		}
	}
	if !validSHA(input.SubjectSHA) || len(input.Assurance) == 0 || len(input.Eligibility) == 0 || len(input.Merge) == 0 {
		state.Reason = ReasonUnavailable
		return state
	}
	var assurance assuranceReport
	var eligibility eligibilityReport
	var merge mergeProof
	if json.Unmarshal(input.Assurance, &assurance) != nil || json.Unmarshal(input.Eligibility, &eligibility) != nil || json.Unmarshal(input.Merge, &merge) != nil {
		state.Reason = ReasonUnavailable
		return state
	}
	state.Eligibility = eligibility
	if reason := unknownReason(assurance, eligibility, merge); reason != "" {
		state.Reason = reason
		return state
	}
	if state.RawExact != len(bindings) {
		state.Reason, state.Resolution = ReasonDigestMismatch, ResolutionInvariant
		return state
	}
	if !validAssurance(assurance) {
		state.Reason, state.Resolution = ReasonAssuranceMismatch, ResolutionInvariant
		return state
	}
	state.AssuranceExact = 1
	if !validEligibility(eligibility) {
		state.Reason, state.Resolution = ReasonEligibilityMismatch, ResolutionInvariant
		return state
	}
	state.EligibilityExact = 1
	if !validMerge(merge) {
		state.Reason, state.Resolution = ReasonMergeMismatch, ResolutionInvariant
		return state
	}
	state.MergeExact = 1
	state.Resolution = ResolutionExact
	return state
}

func unknownReason(assurance assuranceReport, eligibility eligibilityReport, merge mergeProof) string {
	if eligibility.Decision != "ELIGIBLE_SHADOW" || eligibility.Resolution != ResolutionExact {
		return ReasonEligibilityUnknown
	}
	if assurance.AssuranceDecision != "PARTIAL" || assurance.CandidateDecision != "ALLOW_LIMITED" || assurance.CandidateResolution != ResolutionExact {
		return ReasonAssuranceUnknown
	}
	if merge.State != "MERGED" {
		return ReasonMergeUnknown
	}
	return ""
}
