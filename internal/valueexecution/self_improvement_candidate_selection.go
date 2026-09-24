package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementCandidateSelectionObservationSchema = "gooo.self-improvement.candidate-selection.v1"

type SelfImprovementCandidateSelectionStatus string

const SelfImprovementCandidateSelectionStatusSelected SelfImprovementCandidateSelectionStatus = "SELECTED"
const SelfImprovementCandidateSelectionStatusRejected SelfImprovementCandidateSelectionStatus = "REJECTED"
const SelfImprovementCandidateSelectionStatusUnknown SelfImprovementCandidateSelectionStatus = "UNKNOWN"

type SelfImprovementCandidateSelectionObservation struct {
	Schema          string                                  `json:"schema"`
	PromotionDigest string                                  `json:"promotion_digest"`
	ReplayDigest    string                                  `json:"replay_digest"`
	Status          SelfImprovementCandidateSelectionStatus `json:"status"`
	Reason          string                                  `json:"reason"`
	NonAuthorizing  bool                                    `json:"non_authorizing"`
	Digest          string                                  `json:"digest"`
}

// ObserveSelfImprovementCandidateSelection records a candidate selection
// decision without executing or promoting the candidate.
func ObserveSelfImprovementCandidateSelection(
	promotion SelfImprovementPromotionObservation,
	replay SelfImprovementCounterexampleReplayObservation,
) SelfImprovementCandidateSelectionObservation {
	selection := SelfImprovementCandidateSelectionObservation{
		Schema:          SelfImprovementCandidateSelectionObservationSchema,
		PromotionDigest: promotion.Digest,
		ReplayDigest:    replay.Digest,
		Status:          SelfImprovementCandidateSelectionStatusUnknown,
		Reason:          "CANDIDATE_SELECTION_UNKNOWN",
		NonAuthorizing:  true,
	}
	switch {
	case !promotion.NonAuthorizing || !replay.NonAuthorizing || promotion.Digest == "" || replay.Digest == "":
		selection.Reason = "CANDIDATE_SELECTION_EVIDENCE_UNTRUSTED"
	case promotion.Status == SelfImprovementPromotionStatusUnknown || replay.Status == SelfImprovementCounterexampleReplayStatusUnknown:
		selection.Reason = "CANDIDATE_SELECTION_UNKNOWN_INPUT"
	case promotion.Status != SelfImprovementPromotionStatusEligible:
		selection.Status = SelfImprovementCandidateSelectionStatusRejected
		selection.Reason = "CANDIDATE_SELECTION_PROMOTION_NOT_ELIGIBLE"
	case replay.Status != SelfImprovementCounterexampleReplayStatusMatched:
		selection.Status = SelfImprovementCandidateSelectionStatusRejected
		selection.Reason = "CANDIDATE_SELECTION_REPLAY_NOT_MATCHED"
	default:
		selection.Status = SelfImprovementCandidateSelectionStatusSelected
		selection.Reason = "CANDIDATE_SELECTION_EVIDENCE_BOUND"
	}
	selection.Digest = selfImprovementCandidateSelectionObservationDigest(selection)
	return selection
}

func selfImprovementCandidateSelectionObservationDigest(observation SelfImprovementCandidateSelectionObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
