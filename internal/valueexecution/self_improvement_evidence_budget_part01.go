package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const SelfImprovementEvidenceBudgetObservationSchema = "gooo.self-improvement.evidence-budget.v1"

type SelfImprovementEvidenceBudgetStatus string

const (
	SelfImprovementEvidenceBudgetAdequate    SelfImprovementEvidenceBudgetStatus = "ADEQUATE"
	SelfImprovementEvidenceBudgetInsufficient SelfImprovementEvidenceBudgetStatus = "INSUFFICIENT"
	SelfImprovementEvidenceBudgetUnknown     SelfImprovementEvidenceBudgetStatus = "UNKNOWN"
)

type SelfImprovementEvidenceBudgetInput struct {
	ForwardEvidenceCount  int    `json:"forwardEvidenceCount"`
	ReverseEvidenceCount  int    `json:"reverseEvidenceCount"`
	RequiredEvidenceCount int    `json:"requiredEvidenceCount"`
	ForwardDigest         string `json:"forwardDigest"`
	ReverseDigest         string `json:"reverseDigest"`
}

type SelfImprovementEvidenceBudgetObservation struct {
	Schema                string                             `json:"schema"`
	ForwardEvidenceCount  int                                `json:"forwardEvidenceCount"`
	ReverseEvidenceCount  int                                `json:"reverseEvidenceCount"`
	RequiredEvidenceCount int                                `json:"requiredEvidenceCount"`
	Status                SelfImprovementEvidenceBudgetStatus `json:"status"`
	Reason                string                             `json:"reason"`
	NonAuthorizing        bool                               `json:"nonAuthorizing"`
	Digest                string                             `json:"digest"`
}

func ObserveSelfImprovementEvidenceBudget(input SelfImprovementEvidenceBudgetInput) SelfImprovementEvidenceBudgetObservation {
	observation := SelfImprovementEvidenceBudgetObservation{
		Schema:                SelfImprovementEvidenceBudgetObservationSchema,
		ForwardEvidenceCount:  input.ForwardEvidenceCount,
		ReverseEvidenceCount:  input.ReverseEvidenceCount,
		RequiredEvidenceCount: input.RequiredEvidenceCount,
		NonAuthorizing:        true,
	}

	switch {
	case input.RequiredEvidenceCount <= 0:
		observation.Status = SelfImprovementEvidenceBudgetUnknown
		observation.Reason = "REQUIRED_EVIDENCE_COUNT_INVALID"
	case input.ForwardEvidenceCount < 0 || input.ReverseEvidenceCount < 0:
		observation.Status = SelfImprovementEvidenceBudgetUnknown
		observation.Reason = "EVIDENCE_COUNT_INVALID"
	case input.ForwardDigest == "" || input.ReverseDigest == "":
		observation.Status = SelfImprovementEvidenceBudgetUnknown
		observation.Reason = "EVIDENCE_DIGEST_MISSING"
	case input.ForwardEvidenceCount+input.ReverseEvidenceCount < input.RequiredEvidenceCount:
		observation.Status = SelfImprovementEvidenceBudgetInsufficient
		observation.Reason = "EVIDENCE_BUDGET_INSUFFICIENT"
	default:
		observation.Status = SelfImprovementEvidenceBudgetAdequate
		observation.Reason = "EVIDENCE_BUDGET_ADEQUATE"
	}

	observation.Digest = digestSelfImprovementEvidenceBudget(observation)
	return observation
}

func digestSelfImprovementEvidenceBudget(observation SelfImprovementEvidenceBudgetObservation) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}