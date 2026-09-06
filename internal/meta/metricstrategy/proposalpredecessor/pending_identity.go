package proposalpredecessor

import (
	"regexp"
	"slices"
)

var pendingPredecessorID = regexp.MustCompile("^workflow_run:[1-9][0-9]*:attempt:[1-9][0-9]*$")

// AwaitablePending is scheduling information, never selection authority.
func AwaitablePending(report Report) bool {
	if Validate(report) != nil || report.Ready() || report.Reason != ReasonEvidenceUnknown ||
		report.Decision != ResolutionFailClosed || report.ObservationDecision != DecisionUnknown ||
		report.Unknown == nil || report.Unknown.UnknownClass != PendingObservationClass ||
		report.Unknown.NextOperation != PendingObservationOperation {
		return false
	}
	frontier := report.Unknown.BlockedBy
	if len(frontier) == 0 || len(frontier) != report.Summary.UnresolvedCandidates ||
		report.Summary.ExactRuns < len(frontier) ||
		report.Summary.ObservedRuns < report.Summary.ExactRuns || !slices.IsSorted(frontier) {
		return false
	}
	for index, id := range frontier {
		if !pendingPredecessorID.MatchString(id) || (index > 0 && frontier[index-1] == id) {
			return false
		}
	}
	return true
}
