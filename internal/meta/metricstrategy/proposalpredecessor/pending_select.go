package proposalpredecessor

import (
	"fmt"
	"slices"
)

const PendingObservationClass = "DEPENDENCY_BLOCKED"
const PendingObservationOperation = "OBSERVE_PREDECESSOR_WORKFLOW_TERMINAL"

// SelectPending annotates only a directly observed, exact-route dependency.
// Selection, contradiction priority and the five existing proofs stay unchanged.
func SelectPending(repository, currentSHA, predecessorSHA string, collection Collection) (Report, []byte, error) {
	report, payload, err := Select(repository, currentSHA, predecessorSHA, collection)
	if err == nil || !onlyPendingPredecessors(collection, predecessorSHA) {
		return report, payload, err
	}
	frontier := make([]string, 0, len(collection.pending))
	for _, run := range collection.pending {
		frontier = append(frontier, fmt.Sprintf("workflow_run:%d:attempt:%d", run.ID, run.RunAttempt))
	}
	slices.Sort(frontier)
	report.Reason = ReasonEvidenceUnknown
	report.Unknown = &Unknown{
		Stage: ResolutionStage, Step: ResolutionStep, Reason: ReasonEvidenceUnknown,
		UnknownClass: PendingObservationClass, NextOperation: PendingObservationOperation,
		BlockedBy: frontier,
	}
	report, sealErr := sealReport(report)
	if sealErr != nil {
		return Report{}, nil, sealErr
	}
	if validateErr := Validate(report); validateErr != nil {
		return Report{}, nil, validateErr
	}
	return report, nil, fmt.Errorf("proposal predecessor remains dependency-blocked: %s", report.Reason)
}

func onlyPendingPredecessors(collection Collection, predecessorSHA string) bool {
	if len(collection.pending) == 0 || collection.Unresolved != len(collection.pending) ||
		collection.Contradictions != 0 || collection.RouteUnknownRuns != 0 ||
		collection.FailureReason != "" || len(collection.Candidates) > 1 {
		return false
	}
	if len(collection.Candidates) == 1 &&
		(!candidateReady(collection.Candidates[0], predecessorSHA) || collection.ExactJobs != 1) {
		return false
	}
	seen := map[int64]bool{}
	for _, run := range collection.pending {
		if !isInFlightPredecessorRun(run) || run.HeadSHA != predecessorSHA ||
			run.HeadBranch != collection.RequestedRoute || seen[run.ID] {
			return false
		}
		seen[run.ID] = true
	}
	return true
}
