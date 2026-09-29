package lsp

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const (
	ExecutionEvidenceReceiptClosureObservationSchema = provenance.ExecutionEvidenceReceiptClosureObservationSchemaPart01
)

type ExecutionEvidenceReceiptClosureStatus = provenance.ExecutionEvidenceReceiptClosureStatusPart01

const (
	ExecutionEvidenceReceiptClosureComplete = provenance.ExecutionEvidenceReceiptClosureCompletePart01
	ExecutionEvidenceReceiptClosureUnknown  = provenance.ExecutionEvidenceReceiptClosureUnknownPart01
)

type ExecutionEvidenceReceiptClosureObservation = provenance.ExecutionEvidenceReceiptClosureObservationPart01

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
	case !receipt.Validate():
		if receipt.Status == "" && receipt.ReceiptDigest == "" {
			observation.Reason = "EXECUTION_ORIGIN_RECEIPT_MISSING"
		} else {
			observation.Reason = "EXECUTION_ORIGIN_RECEIPT_INVALID"
		}
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
	observation.Digest = observation.ComputedDigest()
	return observation
}

func ValidateExecutionEvidenceReceiptClosureObservation(
	observation ExecutionEvidenceReceiptClosureObservation,
	prefix ExecutionEvidencePrefixObservation,
	receipt valueexecution.ExecutionOriginReceipt,
) bool {
	expected := ObserveExecutionEvidenceReceiptClosure(prefix, receipt)
	return observation.Schema == expected.Schema &&
		observation.Validate() &&
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
	return observation.ComputedDigest()
}
