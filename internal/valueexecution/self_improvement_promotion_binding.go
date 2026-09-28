package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementPromotionBindingObservationSchema = "gooo.self-improvement.promotion-binding-observation.v1"

type SelfImprovementPromotionBindingStatus string

const SelfImprovementPromotionBindingStatusComplete SelfImprovementPromotionBindingStatus = "COMPLETE"
const SelfImprovementPromotionBindingStatusBlocked SelfImprovementPromotionBindingStatus = "BLOCKED"
const SelfImprovementPromotionBindingStatusUnknown SelfImprovementPromotionBindingStatus = "UNKNOWN"

type SelfImprovementPromotionBindingObservation struct {
	Schema          string                                `json:"schema"`
	PromotionDigest string                                `json:"promotion_digest"`
	ReceiptDigest   string                                `json:"receipt_digest"`
	Status          SelfImprovementPromotionBindingStatus `json:"status"`
	Reason          string                                `json:"reason"`
	NonAuthorizing  bool                                  `json:"non_authorizing"`
	Digest          string                                `json:"digest"`
}

// ObserveSelfImprovementPromotionBinding closes the evidence boundary between
// a ledger decision and the execution receipt that produced its observation.
// It is evidence only and never grants promotion, merge, or execution rights.
func ObserveSelfImprovementPromotionBinding(
	promotion SelfImprovementPromotionObservation,
	receipt ExecutionOriginReceipt,
) SelfImprovementPromotionBindingObservation {
	observation := SelfImprovementPromotionBindingObservation{
		Schema:          SelfImprovementPromotionBindingObservationSchema,
		PromotionDigest: promotion.Digest,
		ReceiptDigest:   receipt.ReceiptDigest,
		Status:          SelfImprovementPromotionBindingStatusUnknown,
		Reason:          "PROMOTION_EVIDENCE_UNKNOWN",
		NonAuthorizing:  true,
	}
	switch {
	case promotion.Status == SelfImprovementPromotionStatusUnknown || promotion.Status == "":
		observation.Reason = "PROMOTION_STATUS_UNKNOWN"
	case promotion.Status != SelfImprovementPromotionStatusEligible:
		observation.Status = SelfImprovementPromotionBindingStatusBlocked
		observation.Reason = "PROMOTION_NOT_ELIGIBLE"
	case receipt.Status != ExecutionOriginStatusBound:
		observation.Reason = "EXECUTION_ORIGIN_UNBOUND"
	case receipt.Phase != ExecutionPhaseCompleted:
		observation.Reason = "EXECUTION_NOT_COMPLETED"
	case promotion.Digest == "" || receipt.ReceiptDigest == "" || receipt.ExecutionDigest == "":
		observation.Reason = "PROMOTION_RECEIPT_DIGEST_MISSING"
	default:
		observation.Status = SelfImprovementPromotionBindingStatusComplete
		observation.Reason = "PROMOTION_EXECUTION_BOUND"
	}
	observation.Digest = selfImprovementPromotionBindingObservationDigest(observation)
	return observation
}

func selfImprovementPromotionBindingObservationDigest(observation SelfImprovementPromotionBindingObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
