package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementCycleObservationSchema = "gooo.self-improvement.cycle-observation.v1"

type SelfImprovementCycleStatus string

const SelfImprovementCycleStatusReady SelfImprovementCycleStatus = "READY"
const SelfImprovementCycleStatusPaused SelfImprovementCycleStatus = "PAUSED"
const SelfImprovementCycleStatusUnknown SelfImprovementCycleStatus = "UNKNOWN"

type SelfImprovementCycleObservation struct {
	Schema             string                     `json:"schema"`
	ContinuationDigest string                     `json:"continuation_digest"`
	RequestDigest      string                     `json:"request_digest"`
	Status             SelfImprovementCycleStatus `json:"status"`
	Reason             string                     `json:"reason"`
	NonAuthorizing     bool                       `json:"non_authorizing"`
	Digest             string                     `json:"digest"`
}

// ObserveSelfImprovementCycle closes one evidence-driven cycle boundary. READY
// is a cycle state, not permission to execute, merge, or deploy.
func ObserveSelfImprovementCycle(
	continuation SelfImprovementContinuationObservation,
	request SelfImprovementExecutionRequestObservation,
) SelfImprovementCycleObservation {
	cycle := SelfImprovementCycleObservation{
		Schema:             SelfImprovementCycleObservationSchema,
		ContinuationDigest: continuation.Digest,
		RequestDigest:      request.Digest,
		Status:             SelfImprovementCycleStatusUnknown,
		Reason:             "CYCLE_UNKNOWN",
		NonAuthorizing:     true,
	}
	switch {
	case !continuation.NonAuthorizing || !request.NonAuthorizing || continuation.Digest == "" || request.Digest == "":
		cycle.Reason = "CYCLE_EVIDENCE_UNTRUSTED"
	case continuation.Status == SelfImprovementContinuationStatusUnknown || request.Status == SelfImprovementExecutionRequestStatusUnknown:
		cycle.Reason = "CYCLE_INPUT_UNKNOWN"
	case continuation.Status == SelfImprovementContinuationStatusPause || request.Status == SelfImprovementExecutionRequestStatusRejected:
		cycle.Status = SelfImprovementCycleStatusPaused
		cycle.Reason = "CYCLE_PAUSED_BY_EVIDENCE"
	case continuation.Status != SelfImprovementContinuationStatusContinue:
		cycle.Reason = "CYCLE_CONTINUATION_NOT_READY"
	case request.Status != SelfImprovementExecutionRequestStatusReady:
		cycle.Status = SelfImprovementCycleStatusPaused
		cycle.Reason = "CYCLE_REQUEST_NOT_READY"
	default:
		cycle.Status = SelfImprovementCycleStatusReady
		cycle.Reason = "CYCLE_EVIDENCE_READY"
	}
	cycle.Digest = selfImprovementCycleObservationDigest(cycle)
	return cycle
}

func selfImprovementCycleObservationDigest(observation SelfImprovementCycleObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
