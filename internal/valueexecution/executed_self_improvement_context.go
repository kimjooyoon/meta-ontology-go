package valueexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

// ObserveExecutedSelfImprovementWithContext binds completed execution receipts
// to a comparison whose candidate source identities may differ under one fixed
// evaluation context.
func ObserveExecutedSelfImprovementWithContext(
	beforeOrigin, afterOrigin provenance.OriginChainObservation,
	beforeContext, afterContext ImprovementContextObservation,
	beforeReceipt, afterReceipt ExecutionOriginReceipt,
	beforeMetric, afterMetric ImprovementMetric,
) ExecutedSelfImprovementObservation {
	candidate := ObserveSelfImprovementWithContext(
		beforeOrigin, afterOrigin, beforeContext, afterContext,
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
