package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const ExecutionEvidenceReceiptClosureObservationSchemaPart01 = "gooo.lsp.execution-evidence-receipt-closure.v1"

type ExecutionEvidenceReceiptClosureStatusPart01 string

const (
	ExecutionEvidenceReceiptClosureCompletePart01 ExecutionEvidenceReceiptClosureStatusPart01 = "COMPLETE"
	ExecutionEvidenceReceiptClosureUnknownPart01  ExecutionEvidenceReceiptClosureStatusPart01 = "UNKNOWN"
)

// ExecutionEvidenceReceiptClosureObservationPart01 is the non-authorizing
// closure summary attached to an execution-plan provenance response.
type ExecutionEvidenceReceiptClosureObservationPart01 struct {
	Schema               string                                      `json:"schema"`
	EvidencePrefixDigest string                                      `json:"evidence_prefix_digest"`
	ReceiptDigest        string                                      `json:"receipt_digest"`
	MissingStageIndex    int                                         `json:"missing_stage_index"`
	Status               ExecutionEvidenceReceiptClosureStatusPart01 `json:"status"`
	Reason               string                                      `json:"reason"`
	NonAuthorizing       bool                                        `json:"non_authorizing"`
	Digest               string                                      `json:"digest"`
}

func (observation ExecutionEvidenceReceiptClosureObservationPart01) ComputedDigest() string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (observation ExecutionEvidenceReceiptClosureObservationPart01) Validate() bool {
	if observation.Schema != ExecutionEvidenceReceiptClosureObservationSchemaPart01 ||
		observation.EvidencePrefixDigest == "" || observation.MissingStageIndex < -1 ||
		observation.Reason == "" || !observation.NonAuthorizing ||
		observation.Digest == "" || observation.Digest != observation.ComputedDigest() {
		return false
	}
	switch observation.Status {
	case ExecutionEvidenceReceiptClosureCompletePart01:
		return observation.MissingStageIndex == -1 && observation.ReceiptDigest != "" &&
			observation.Reason == "EXECUTION_EVIDENCE_RECEIPT_CLOSED"
	case ExecutionEvidenceReceiptClosureUnknownPart01:
		return observation.Reason != "EXECUTION_EVIDENCE_RECEIPT_CLOSED"
	default:
		return false
	}
}
