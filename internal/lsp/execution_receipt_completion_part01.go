package lsp

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const ExecutionBackedCompletionObservationSchema = "gooo.lsp.execution-backed-completion.v1"

type ExecutionBackedCompletionStatus string

const ExecutionBackedCompletionStatusComplete ExecutionBackedCompletionStatus = "COMPLETE"
const ExecutionBackedCompletionStatusUnknown ExecutionBackedCompletionStatus = "UNKNOWN"

type ExecutionBackedCompletionObservation struct {
	Schema           string                          `json:"schema"`
	CompletionDigest string                          `json:"completion_digest"`
	ReceiptDigest    string                          `json:"receipt_digest"`
	Status           ExecutionBackedCompletionStatus `json:"status"`
	Reason           string                          `json:"reason"`
	NonAuthorizing   bool                            `json:"non_authorizing"`
	Digest           string                          `json:"digest"`
}

// ObserveExecutionBackedCompletion connects editor evidence to a completed
// runtime receipt without making a completion item executable or trusted.
func ObserveExecutionBackedCompletion(
	context DeclarationCompletionContext,
	completion DeclarationCompletionObservation,
	receipt valueexecution.ExecutionOriginReceipt,
) ExecutionBackedCompletionObservation {
	observation := ExecutionBackedCompletionObservation{
		Schema:           ExecutionBackedCompletionObservationSchema,
		CompletionDigest: completion.Digest,
		ReceiptDigest:    receipt.ReceiptDigest,
		Status:           ExecutionBackedCompletionStatusUnknown,
		Reason:           "COMPLETION_PROVENANCE_INCOMPLETE",
		NonAuthorizing:   true,
	}
	switch {
	case completion.Status != DeclarationCompletionStatusComplete:
		observation.Reason = "COMPLETION_PROVENANCE_INCOMPLETE"
	case completion.DocumentURI != context.DocumentURI ||
		completion.DocumentVersion != context.DocumentVersion ||
		completion.DeclarationSymbol != context.DeclarationSymbol ||
		completion.EnvironmentDigest != context.EnvironmentDigest:
		observation.Reason = "COMPLETION_CONTEXT_MISMATCH"
	case receipt.Status != valueexecution.ExecutionOriginStatusBound:
		observation.Reason = "EXECUTION_ORIGIN_UNBOUND"
	case receipt.Phase != valueexecution.ExecutionPhaseCompleted:
		observation.Reason = "EXECUTION_NOT_COMPLETED"
	case completion.Digest == "" || receipt.ReceiptDigest == "" || receipt.ExecutionDigest == "":
		observation.Reason = "COMPLETION_RECEIPT_DIGEST_MISSING"
	default:
		observation.Status = ExecutionBackedCompletionStatusComplete
		observation.Reason = "COMPLETION_EXECUTION_BOUND"
	}
	observation.Digest = executionBackedCompletionObservationDigest(observation)
	return observation
}

func executionBackedCompletionObservationDigest(observation ExecutionBackedCompletionObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
