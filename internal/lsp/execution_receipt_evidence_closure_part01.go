package lsp

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const ExecutionEvidenceReceiptClosureObservationSchema = "gooo.lsp.execution-evidence-receipt-closure.v1"

type ExecutionEvidenceReceiptClosureStatus string

const (
	ExecutionEvidenceReceiptClosureComplete ExecutionEvidenceReceiptClosureStatus = "COMPLETE"
	ExecutionEvidenceReceiptClosureUnknown  ExecutionEvidenceReceiptClosureStatus = "UNKNOWN"
)

type ExecutionEvidenceReceiptClosureObservation struct {
	Schema               string                                `json:"schema"`
	EvidencePrefixDigest string                                `json:"evidence_prefix_digest"`
	ReceiptDigest        string                                `json:"receipt_digest"`
	MissingStageIndex    int                                   `json:"missing_stage_index"`
	Status               ExecutionEvidenceReceiptClosureStatus `json:"status"`
	Reason               string                                `json:"reason"`
	NonAuthorizing       bool                                  `json:"non_authorizing"`
	Digest               string                                `json:"digest"`
}

// ObserveExecutionEvidenceReceiptClosure keeps LSP completion non-authorizing
// until the entire observed evidence prefix and runtime receipt are complete.
func ObserveExecutionEvidenceReceiptClosure(
	prefix ExecutionEvidencePrefixObservation,
	receipt valueexecution.ExecutionOriginReceipt,
) ExecutionEvidenceReceiptClosureObservation {
	observation := ExecutionEvidenceReceiptClosureObservation{
		Schema:               ExecutionEvidenceReceiptClosureObservationSchema,
		EvidencePrefixDigest: prefix.EvidencePrefixDigest,
		ReceiptDigest:        receipt.ReceiptDigest,
		MissingStageIndex:    prefix.MissingStageIndex,
		Status:               ExecutionEvidenceReceiptClosureUnknown,
		Reason:               "EXECUTION_EVIDENCE_PREFIX_INCOMPLETE",
		NonAuthorizing:       true,
	}
	switch {
	case !ValidateExecutionEvidencePrefixObservation(prefix):
		observation.Reason = "EXECUTION_EVIDENCE_PREFIX_INVALID"
	case prefix.Status != ExecutionEvidencePrefixComplete:
		observation.Reason = "EXECUTION_EVIDENCE_PREFIX_INCOMPLETE"
	case prefix.EvidencePrefixDigest == "":
		observation.Reason = "EXECUTION_EVIDENCE_PREFIX_DIGEST_MISSING"
	case receipt.Status != valueexecution.ExecutionOriginStatusBound:
		observation.Reason = "EXECUTION_ORIGIN_UNBOUND"
	case receipt.Phase != valueexecution.ExecutionPhaseCompleted:
		observation.Reason = "EXECUTION_NOT_COMPLETED"
	case receipt.ReceiptDigest == "" || receipt.ExecutionDigest == "":
		observation.Reason = "EXECUTION_RECEIPT_DIGEST_MISSING"
	default:
		observation.Status = ExecutionEvidenceReceiptClosureComplete
		observation.Reason = "EXECUTION_EVIDENCE_RECEIPT_CLOSED"
	}
	observation.Digest = executionEvidenceReceiptClosureObservationDigest(observation)
	return observation
}

func ValidateExecutionEvidenceReceiptClosureObservation(
	observation ExecutionEvidenceReceiptClosureObservation,
	prefix ExecutionEvidencePrefixObservation,
	receipt valueexecution.ExecutionOriginReceipt,
) bool {
	expected := ObserveExecutionEvidenceReceiptClosure(prefix, receipt)
	return observation.Schema == expected.Schema &&
		observation.EvidencePrefixDigest == expected.EvidencePrefixDigest &&
		observation.ReceiptDigest == expected.ReceiptDigest &&
		observation.MissingStageIndex == expected.MissingStageIndex &&
		observation.Status == expected.Status &&
		observation.Reason == expected.Reason &&
		observation.NonAuthorizing &&
		observation.Digest == expected.Digest
}

func executionEvidenceReceiptClosureObservationDigest(
	observation ExecutionEvidenceReceiptClosureObservation,
) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
