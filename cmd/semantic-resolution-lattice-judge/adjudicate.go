package main

import (
	"errors"
	"fmt"
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

func reconstructCase(item declaredCase) latticeCase {
	result := adjudicate(item.Observation)
	return latticeCase{
		ID: item.ID, Decision: caseDecision(result), Observation: item.Observation,
		Transition: result, ClaimID: item.ClaimID,
	}
}

func reconstructClaim(item declaredCase, result transition) claim {
	before := item.ClaimPriorState
	after := deriveClaimAfterState(before, result.Decision)
	stage, step, reason := claimEvidenceFields(result)
	return claim{ID: item.ClaimID, State: after, BeforeState: before, AfterState: after,
		Preserved: before == after, Stage: stage, Step: step, Reason: reason,
		EvidenceDigest: claimEvidenceDigest(item.ClaimID, before, after, item.Observation, result),
		Provenance:     "gooo://semantic-resolution-lattice/case/" + item.ID}
}

func deriveClaimAfterState(before, decision string) string {
	switch decision {
	case "PASS":
		if before == "OPEN" {
			return "DISCHARGED"
		}
	case "LOWER_RESOLUTION", "UNKNOWN":
		return before
	case "FAIL_CLOSED":
		if before != "REFUTED" {
			return "REFUTED"
		}
	}
	return before
}

func claimEvidenceFields(result transition) (string, int, string) {
	if result.Unknown != nil {
		return result.Unknown.Stage, result.Unknown.Step, result.Unknown.Reason
	}
	if result.Decision == "PASS" {
		return "EXACT", 0, result.Reason
	}
	return "FAIL_CLOSED", 1, result.Reason
}

func claimEvidenceDigest(claimID, before, after string, input observation, result transition) string {
	unknownStage, unknownStep, unknownReason := "", 0, ""
	if result.Unknown != nil {
		unknownStage, unknownStep, unknownReason = result.Unknown.Stage, result.Unknown.Step, result.Unknown.Reason
	}
	canonical := fmt.Sprintf("%s|%s|%s|%d|%d|%s|%d|%t|%s|%s|%s|%s|%s|%d|%s\n",
		claimID, before, after, input.Required, input.Observed, input.Reason,
		input.RepositoryWrites, input.MutationAuthority, result.FromResolution,
		result.ToResolution, result.Decision, result.Reason, unknownStage, unknownStep, unknownReason)
	return digestText(canonical)
}

func caseDecision(result transition) string {
	if result.Decision == "LOWER_RESOLUTION" {
		return "UNKNOWN"
	}
	return result.Decision
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
