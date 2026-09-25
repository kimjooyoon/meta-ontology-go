package valueexecution

import (
	"encoding/json"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

// ExecutionOriginStatus describes an observation binding, not execution
// authority or semantic success.
type ExecutionOriginStatus string

const (
	ExecutionOriginStatusUnknown ExecutionOriginStatus = "UNKNOWN"
	ExecutionOriginStatusBound   ExecutionOriginStatus = "BOUND"
)

// ExecutionOriginReceipt links a complete origin observation to one detached
// execution summary. Failed and in-flight phases remain visible as-is.
type ExecutionOriginReceipt struct {
	OriginDigest      string                `json:"origin_digest,omitempty"`
	ExecutionDigest   string                `json:"execution_digest,omitempty"`
	RuntimePlanDigest string                `json:"runtime_plan_digest,omitempty"`
	Phase             ExecutionPhase        `json:"phase"`
	Conditions        []ExecutionCondition  `json:"conditions"`
	Status            ExecutionOriginStatus `json:"status"`
	Reason            string                `json:"reason"`
	NonAuthorizing    bool                  `json:"non_authorizing"`
	ReceiptDigest     string                `json:"receipt_digest"`
}

// ObserveExecutionOrigin creates a deterministic, non-authorizing link
// between provenance and an execution receipt. It refuses incomplete origin
// evidence and never turns a failed phase into a successful observation.
func ObserveExecutionOrigin(origin provenance.OriginChainObservation, execution Execution) ExecutionOriginReceipt {
	receipt := ExecutionOriginReceipt{
		OriginDigest:      origin.Digest,
		ExecutionDigest:   execution.ExecutionDigest,
		RuntimePlanDigest: execution.RuntimePlanDigest,
		Phase:             execution.Phase,
		Conditions:        canonicalOriginExecutionConditions(execution.Conditions),
		Status:            ExecutionOriginStatusUnknown,
		NonAuthorizing:    true,
	}
	switch {
	case !origin.Verified() || !origin.Comparable():
		receipt.Reason = "ORIGIN_OBSERVATION_INCOMPLETE"
	case execution.ExecutionDigest == "":
		receipt.Reason = "EXECUTION_DIGEST_MISSING"
	case !knownExecutionPhase(execution.Phase):
		receipt.Reason = "EXECUTION_PHASE_UNKNOWN"
	default:
		receipt.Status = ExecutionOriginStatusBound
		receipt.Reason = "ORIGIN_BOUND_TO_EXECUTION"
	}
	receipt.ReceiptDigest = executionOriginReceiptDigest(receipt)
	return receipt
}

func knownExecutionPhase(phase ExecutionPhase) bool {
	switch phase {
	case ExecutionPhaseRunning, ExecutionPhaseCompleted, ExecutionPhaseFailed:
		return true
	default:
		return false
	}
}

func canonicalOriginExecutionConditions(conditions []ExecutionCondition) []ExecutionCondition {
	result := append([]ExecutionCondition(nil), conditions...)
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].Type != result[right].Type {
			return result[left].Type < result[right].Type
		}
		if result[left].Status != result[right].Status {
			return result[left].Status < result[right].Status
		}
		if result[left].Reason != result[right].Reason {
			return result[left].Reason < result[right].Reason
		}
		return result[left].Detail < result[right].Detail
	})
	return result
}

func executionOriginReceiptDigest(receipt ExecutionOriginReceipt) string {
	receipt.ReceiptDigest = ""
	payload, _ := json.Marshal(receipt)
	return digestBytes(payload)
}
