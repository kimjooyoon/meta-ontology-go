package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementContinuationObservationSchema = "gooo.self-improvement.continuation-observation.v1"

type SelfImprovementContinuationStatus string

const SelfImprovementContinuationStatusContinue SelfImprovementContinuationStatus = "CONTINUE"
const SelfImprovementContinuationStatusPause SelfImprovementContinuationStatus = "PAUSE"
const SelfImprovementContinuationStatusUnknown SelfImprovementContinuationStatus = "UNKNOWN"

type SelfImprovementContinuationObservation struct {
	Schema          string                            `json:"schema"`
	HistoryDigest   string                            `json:"history_digest"`
	SelectionDigest string                            `json:"selection_digest"`
	Status          SelfImprovementContinuationStatus `json:"status"`
	Reason          string                            `json:"reason"`
	NonAuthorizing  bool                              `json:"non_authorizing"`
	Digest          string                            `json:"digest"`
}

// ObserveSelfImprovementContinuation combines prior execution history with a
// candidate selection observation. It never starts or authorizes execution.
func ObserveSelfImprovementContinuation(
	history SelfImprovementExecutionHistoryObservation,
	selection SelfImprovementCandidateSelectionObservation,
) SelfImprovementContinuationObservation {
	continuation := SelfImprovementContinuationObservation{
		Schema:          SelfImprovementContinuationObservationSchema,
		HistoryDigest:   history.Digest,
		SelectionDigest: selection.Digest,
		Status:          SelfImprovementContinuationStatusUnknown,
		Reason:          "CONTINUATION_UNKNOWN",
		NonAuthorizing:  true,
	}
	switch {
	case !history.NonAuthorizing || !selection.NonAuthorizing || history.Digest == "" || selection.Digest == "":
		continuation.Reason = "CONTINUATION_EVIDENCE_UNTRUSTED"
	case history.Decision == SelfImprovementExecutionHistoryDecisionUnknown || selection.Status == SelfImprovementCandidateSelectionStatusUnknown:
		continuation.Reason = "CONTINUATION_INPUT_UNKNOWN"
	case history.Decision == SelfImprovementExecutionHistoryDecisionBlockedRejected || selection.Status == SelfImprovementCandidateSelectionStatusRejected:
		continuation.Status = SelfImprovementContinuationStatusPause
		continuation.Reason = "CONTINUATION_BLOCKED_BY_HISTORY"
	case history.Decision != SelfImprovementExecutionHistoryDecisionCompletedEvidence:
		continuation.Reason = "CONTINUATION_REQUIRES_COMPLETED_HISTORY"
	case selection.Status != SelfImprovementCandidateSelectionStatusSelected:
		continuation.Status = SelfImprovementContinuationStatusPause
		continuation.Reason = "CONTINUATION_CANDIDATE_NOT_SELECTED"
	default:
		continuation.Status = SelfImprovementContinuationStatusContinue
		continuation.Reason = "CONTINUATION_EVIDENCE_BOUND"
	}
	continuation.Digest = selfImprovementContinuationObservationDigest(continuation)
	return continuation
}

func selfImprovementContinuationObservationDigest(observation SelfImprovementContinuationObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
