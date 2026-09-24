package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementCycleResultObservationSchema = "gooo.self-improvement.cycle-result.v1"

type SelfImprovementCycleResultStatus string

const SelfImprovementCycleResultStatusCompleted SelfImprovementCycleResultStatus = "COMPLETED"
const SelfImprovementCycleResultStatusNotApplied SelfImprovementCycleResultStatus = "NOT_APPLIED"
const SelfImprovementCycleResultStatusUnknown SelfImprovementCycleResultStatus = "UNKNOWN"

type SelfImprovementCycleResultObservation struct {
	Schema             string                           `json:"schema"`
	CycleDigest        string                           `json:"cycle_digest"`
	OutcomeDigest      string                           `json:"outcome_digest"`
	RequestDigest      string                           `json:"request_digest"`
	ReceiptDigest      string                           `json:"receipt_digest"`
	Status             SelfImprovementCycleResultStatus `json:"status"`
	Reason             string                           `json:"reason"`
	NonAuthorizing     bool                             `json:"non_authorizing"`
	ReverseObservation bool                             `json:"reverse_observation"`
	EvidenceCount      int                              `json:"evidence_count"`
	Digest             string                           `json:"digest"`
}

// ObserveSelfImprovementCycleResult closes the cycle's forward observation
// with a reverse observation from runtime outcome evidence. It never grants
// permission to execute, merge, or deploy a candidate.
func ObserveSelfImprovementCycleResult(
	cycle SelfImprovementCycleObservation,
	outcome SelfImprovementExecutionOutcomeObservation,
) SelfImprovementCycleResultObservation {
	result := SelfImprovementCycleResultObservation{
		Schema:             SelfImprovementCycleResultObservationSchema,
		CycleDigest:        cycle.Digest,
		OutcomeDigest:      outcome.Digest,
		RequestDigest:      outcome.RequestDigest,
		ReceiptDigest:      outcome.ReceiptDigest,
		Status:             SelfImprovementCycleResultStatusUnknown,
		Reason:             "CYCLE_RESULT_UNKNOWN",
		NonAuthorizing:     true,
		ReverseObservation: true,
	}
	for _, evidence := range []string{result.CycleDigest, result.OutcomeDigest, result.RequestDigest, result.ReceiptDigest} {
		if evidence != "" {
			result.EvidenceCount++
		}
	}
	switch {
	case !cycle.NonAuthorizing || !outcome.NonAuthorizing || cycle.Digest == "" || outcome.Digest == "":
		result.Reason = "CYCLE_RESULT_EVIDENCE_UNTRUSTED"
	case string(cycle.Status) == "UNKNOWN" || outcome.Status == SelfImprovementExecutionOutcomeStatusUnknown:
		result.Reason = "CYCLE_RESULT_INPUT_UNKNOWN"
	case string(cycle.Status) == "PAUSED" || outcome.Status != SelfImprovementExecutionOutcomeStatusCompleted:
		result.Status = SelfImprovementCycleResultStatusNotApplied
		result.Reason = "CYCLE_RESULT_NOT_APPLIED"
	default:
		result.Status = SelfImprovementCycleResultStatusCompleted
		result.Reason = "CYCLE_RESULT_REVERSE_OBSERVED"
	}
	result.Digest = selfImprovementCycleResultObservationDigest(result)
	return result
}

func selfImprovementCycleResultObservationDigest(observation SelfImprovementCycleResultObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
