package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const SelfImprovementContractPreservationObservationSchema = "gooo.self-improvement.contract-preservation.v1"

const (
	SelfImprovementContractPreservationAccepted  = "ACCEPTED"
	SelfImprovementContractPreservationRejected  = "REJECTED"
	SelfImprovementContractPreservationUnknown   = "UNKNOWN"
)

type SelfImprovementContractPreservationInput struct {
	Scope                    string `json:"scope"`
	CounterexampleRecovered  bool   `json:"counterexampleRecovered"`
	BaselineContractDigest   string `json:"baselineContractDigest"`
	CandidateContractDigest  string `json:"candidateContractDigest"`
	CandidateDigest           string `json:"candidateDigest"`
	RegressionEvidenceDigest  string `json:"regressionEvidenceDigest"`
	RegressionStatus          string `json:"regressionStatus"`
}

type SelfImprovementContractPreservationObservation struct {
	Schema                   string `json:"schema"`
	Scope                    string `json:"scope"`
	Status                   string `json:"status"`
	Reason                   string `json:"reason"`
	CounterexampleRecovered  bool   `json:"counterexampleRecovered"`
	ContractPreserved        bool   `json:"contractPreserved"`
	RegressionEvidenceSeen   bool   `json:"regressionEvidenceSeen"`
	AdoptionAuthorized       bool   `json:"adoptionAuthorized"`
	NonAuthorizing           bool   `json:"nonAuthorizing"`
	Digest                   string `json:"digest"`
}

func ObserveSelfImprovementContractPreservation(
	input SelfImprovementContractPreservationInput,
) SelfImprovementContractPreservationObservation {
	observation := SelfImprovementContractPreservationObservation{
		Schema:                  SelfImprovementContractPreservationObservationSchema,
		Scope:                   input.Scope,
		CounterexampleRecovered: input.CounterexampleRecovered,
		AdoptionAuthorized:      false,
		NonAuthorizing:          true,
	}

	switch {
	case input.Scope == "" || input.CandidateDigest == "":
		observation.Status = SelfImprovementContractPreservationUnknown
		observation.Reason = "CONTRACT_SCOPE_OR_CANDIDATE_MISSING"
	case !input.CounterexampleRecovered:
		observation.Status = SelfImprovementContractPreservationUnknown
		observation.Reason = "COUNTEREXAMPLE_NOT_RECOVERED"
	case input.BaselineContractDigest == "" || input.CandidateContractDigest == "":
		observation.Status = SelfImprovementContractPreservationUnknown
		observation.Reason = "CONTRACT_DIGEST_MISSING"
	case input.BaselineContractDigest != input.CandidateContractDigest:
		observation.Status = SelfImprovementContractPreservationRejected
		observation.Reason = "CONTRACT_CHANGED_REQUIRES_EXPLICIT_SCOPE"
	case input.RegressionStatus == "FAILED":
		observation.Status = SelfImprovementContractPreservationRejected
		observation.Reason = "CONTRACT_REGRESSION_DETECTED"
	case input.RegressionEvidenceDigest == "" || input.RegressionStatus != "PASSED":
		observation.Status = SelfImprovementContractPreservationUnknown
		observation.Reason = "REGRESSION_EVIDENCE_MISSING_OR_UNKNOWN"
	default:
		observation.Status = SelfImprovementContractPreservationAccepted
		observation.Reason = "CONTRACT_PRESERVED"
		observation.ContractPreserved = true
		observation.RegressionEvidenceSeen = true
	}

	observation.Digest = digestSelfImprovementContractPreservation(observation)
	return observation
}

func digestSelfImprovementContractPreservation(
	observation SelfImprovementContractPreservationObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}