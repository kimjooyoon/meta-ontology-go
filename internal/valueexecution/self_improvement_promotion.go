package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementPromotionObservationSchema = "gooo.self-improvement.promotion-observation.v1"

type SelfImprovementPromotionStatus string

const SelfImprovementPromotionStatusEligible SelfImprovementPromotionStatus = "ELIGIBLE"
const SelfImprovementPromotionStatusBlocked SelfImprovementPromotionStatus = "BLOCKED"
const SelfImprovementPromotionStatusUnknown SelfImprovementPromotionStatus = "UNKNOWN"
const SelfImprovementPromotionStatusInsufficientEvidence SelfImprovementPromotionStatus = "INSUFFICIENT_EVIDENCE"

type SelfImprovementPromotionObservation struct {
	Schema         string                         `json:"schema"`
	LedgerDigest   string                         `json:"ledger_digest"`
	Decision       string                         `json:"decision"`
	Status         SelfImprovementPromotionStatus `json:"status"`
	Reason         string                         `json:"reason"`
	NonAuthorizing bool                           `json:"non_authorizing"`
	Digest         string                         `json:"digest"`
}

// ObserveSelfImprovementPromotion turns ledger state into durable evidence.
// It never authorizes a release, merge, or execution by itself.
func ObserveSelfImprovementPromotion(ledger SelfImprovementEvidenceLedger) SelfImprovementPromotionObservation {
	encoded, _ := json.Marshal(ledger)
	ledgerObservation := ledger.Observe()
	decision := string(ledgerObservation.Decision)
	ledgerDigest := ledger.Digest
	if ledgerDigest == "" {
		ledgerDigest = cache.HashBytes(encoded).String()
	}
	observation := SelfImprovementPromotionObservation{
		Schema:         SelfImprovementPromotionObservationSchema,
		LedgerDigest:   ledgerDigest,
		Decision:       decision,
		Status:         SelfImprovementPromotionStatusInsufficientEvidence,
		Reason:         "LEDGER_NOT_PROMOTION_ELIGIBLE",
		NonAuthorizing: true,
	}
	switch decision {
	case "PROMOTION_ELIGIBLE":
		observation.Status = SelfImprovementPromotionStatusEligible
		observation.Reason = "LEDGER_PROMOTION_ELIGIBLE"
	case "BLOCKED_REGRESSION":
		observation.Status = SelfImprovementPromotionStatusBlocked
		observation.Reason = "LEDGER_BLOCKED_REGRESSION"
	case "UNKNOWN":
		observation.Status = SelfImprovementPromotionStatusUnknown
		observation.Reason = "LEDGER_UNKNOWN"
	case "":
		observation.Reason = "LEDGER_DECISION_MISSING"
	}
	observation.Digest = selfImprovementPromotionObservationDigest(observation)
	return observation
}

func selfImprovementPromotionObservationDigest(observation SelfImprovementPromotionObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
