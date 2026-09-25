package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementExecutionOutcomeObservationSchema = "gooo.self-improvement.execution-outcome.v1"

type SelfImprovementExecutionOutcomeStatus string

const SelfImprovementExecutionOutcomeStatusCompleted SelfImprovementExecutionOutcomeStatus = "COMPLETED"
const SelfImprovementExecutionOutcomeStatusRejected SelfImprovementExecutionOutcomeStatus = "REJECTED"
const SelfImprovementExecutionOutcomeStatusUnknown SelfImprovementExecutionOutcomeStatus = "UNKNOWN"

type SelfImprovementExecutionOutcomeObservation struct {
	Schema         string                                `json:"schema"`
	RequestDigest  string                                `json:"request_digest"`
	ReceiptDigest  string                                `json:"receipt_digest"`
	Status         SelfImprovementExecutionOutcomeStatus `json:"status"`
	Reason         string                                `json:"reason"`
	NonAuthorizing bool                                  `json:"non_authorizing"`
	Digest         string                                `json:"digest"`
}

// ObserveSelfImprovementExecutionOutcome closes a READY request with runtime
// evidence. It never grants permission to execute or deploy a candidate.
func ObserveSelfImprovementExecutionOutcome(
	request SelfImprovementExecutionRequestObservation,
	receipt ExecutionOriginReceipt,
) SelfImprovementExecutionOutcomeObservation {
	outcome := SelfImprovementExecutionOutcomeObservation{
		Schema:         SelfImprovementExecutionOutcomeObservationSchema,
		RequestDigest:  request.Digest,
		ReceiptDigest:  receipt.ReceiptDigest,
		Status:         SelfImprovementExecutionOutcomeStatusUnknown,
		Reason:         "EXECUTION_OUTCOME_UNKNOWN",
		NonAuthorizing: true,
	}
	switch {
	case !request.NonAuthorizing || request.Digest == "":
		outcome.Reason = "EXECUTION_OUTCOME_REQUEST_UNTRUSTED"
	case request.Status == SelfImprovementExecutionRequestStatusUnknown:
		outcome.Reason = "EXECUTION_OUTCOME_REQUEST_UNKNOWN"
	case request.Status != SelfImprovementExecutionRequestStatusReady:
		outcome.Status = SelfImprovementExecutionOutcomeStatusRejected
		outcome.Reason = "EXECUTION_OUTCOME_REQUEST_NOT_READY"
	case receipt.Status != ExecutionOriginStatusBound:
		outcome.Reason = "EXECUTION_OUTCOME_RECEIPT_UNBOUND"
	case receipt.Phase != ExecutionPhaseCompleted:
		outcome.Reason = "EXECUTION_OUTCOME_NOT_COMPLETED"
	case receipt.ReceiptDigest == "" || receipt.ExecutionDigest == "":
		outcome.Reason = "EXECUTION_OUTCOME_DIGEST_MISSING"
	default:
		outcome.Status = SelfImprovementExecutionOutcomeStatusCompleted
		outcome.Reason = "EXECUTION_OUTCOME_COMPLETED"
	}
	outcome.Digest = selfImprovementExecutionOutcomeObservationDigest(outcome)
	return outcome
}

func selfImprovementExecutionOutcomeObservationDigest(observation SelfImprovementExecutionOutcomeObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
