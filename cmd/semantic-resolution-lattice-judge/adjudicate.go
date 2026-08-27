package main

import (
	"errors"
	"reflect"
)

const (
	exact         = "exact_operation"
	invariantOnly = "invariant_only"
)

func adjudicate(input observation) transition {
	result := transition{FromResolution: exact, RepositoryWrites: input.RepositoryWrites, MutationAuthority: input.MutationAuthority}
	switch {
	case input.RepositoryWrites != 0:
		result.Decision, result.Reason = "FAIL_CLOSED", "REPOSITORY_WRITE_EFFECT"
	case input.MutationAuthority:
		result.Decision, result.Reason = "FAIL_CLOSED", "MUTATION_AUTHORITY_PRESENT"
	case input.Required <= 0 || input.Observed < 0 || input.Observed > input.Required:
		result.Decision, result.Reason = "FAIL_CLOSED", "OBSERVATION_CARDINALITY_INVALID"
	case input.Observed == input.Required:
		result.ToResolution, result.Decision, result.Reason = exact, "PASS", "OBSERVATION_COMPLETE"
	default:
		result.ToResolution, result.Decision, result.Reason = invariantOnly, "LOWER_RESOLUTION", "PARTIAL_OBSERVATION"
		result.Unknown = &unknownValue{Stage: "PARTIAL_OBSERVATION", Step: 1, Reason: input.Reason}
	}
	return result
}

func validateCase(item latticeCase) error {
	if item.ID == "" || item.ClaimID == "" || item.Transition.FromResolution != exact {
		return errors.New("invalid case identity")
	}
	if !reflect.DeepEqual(adjudicate(item.Observation), item.Transition) {
		return errors.New("independent transition replay disagrees")
	}
	wantDecision := item.Transition.Decision
	if wantDecision == "LOWER_RESOLUTION" {
		wantDecision = "UNKNOWN"
	}
	if item.Decision != wantDecision {
		return errors.New("case decision is not derived from transition")
	}
	return nil
}
