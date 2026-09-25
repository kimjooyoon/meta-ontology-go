package valueexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

// ExecutedSelfImprovementObservation adds the execution lifecycle boundary to
// a metric comparison. It keeps in-flight and failed receipts visible without
// treating either one as evidence of improvement.
type ExecutedSelfImprovementObservation struct {
	Schema              string                     `json:"schema"`
	Candidate           SelfImprovementObservation `json:"candidate"`
	CandidateDigest     string                     `json:"candidate_digest,omitempty"`
	BeforeReceiptDigest string                     `json:"before_receipt_digest,omitempty"`
	AfterReceiptDigest  string                     `json:"after_receipt_digest,omitempty"`
	Status              SelfImprovementStatus      `json:"status"`
	Reason              string                     `json:"reason"`
	NonAuthorizing      bool                       `json:"non_authorizing"`
	Digest              string                     `json:"digest"`
}

const ExecutedSelfImprovementObservationSchema = "gooo/self-improvement-executed/v1"

// ObserveExecutedSelfImprovement accepts a metric direction only when both
// source-backed execution receipts are complete and bound to their origins.
func ObserveExecutedSelfImprovement(
	beforeOrigin, afterOrigin provenance.OriginChainObservation,
	beforeEnvironment, afterEnvironment EnvironmentObservation,
	beforeReceipt, afterReceipt ExecutionOriginReceipt,
	beforeMetric, afterMetric ImprovementMetric,
) ExecutedSelfImprovementObservation {
	candidate := ObserveSelfImprovement(
		beforeOrigin, afterOrigin, beforeEnvironment, afterEnvironment,
		beforeMetric, afterMetric,
	)
	observation := ExecutedSelfImprovementObservation{
		Schema:              ExecutedSelfImprovementObservationSchema,
		Candidate:           candidate,
		CandidateDigest:     candidate.Digest,
		BeforeReceiptDigest: beforeReceipt.ReceiptDigest,
		AfterReceiptDigest:  afterReceipt.ReceiptDigest,
		Status:              SelfImprovementStatusUnknown,
		NonAuthorizing:      true,
	}
	switch {
	case beforeReceipt.Status != ExecutionOriginStatusBound || afterReceipt.Status != ExecutionOriginStatusBound:
		observation.Reason = "EXECUTION_ORIGIN_UNBOUND"
	case beforeReceipt.Phase != ExecutionPhaseCompleted || afterReceipt.Phase != ExecutionPhaseCompleted:
		observation.Reason = "EXECUTION_NOT_COMPLETED"
	case beforeReceipt.ExecutionDigest == "" || afterReceipt.ExecutionDigest == "":
		observation.Reason = "EXECUTION_DIGEST_MISSING"
	case beforeReceipt.OriginDigest != beforeOrigin.Digest || afterReceipt.OriginDigest != afterOrigin.Digest:
		observation.Reason = "EXECUTION_ORIGIN_MISMATCH"
	default:
		observation.Status = candidate.Status
		observation.Reason = candidate.Reason
	}
	observation.Digest = executedSelfImprovementObservationDigest(observation)
	return observation
}

func executedSelfImprovementObservationDigest(observation ExecutedSelfImprovementObservation) string {
	observation.Digest = ""
	return digestValue(observation)
}
