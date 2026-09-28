package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"time"
)

// ExecutionSuspensionSource is the .gooo declaration that names this
// lifecycle boundary. It is provenance metadata, not an execution capability.
const ExecutionSuspensionSource = "examples/language-runtime-binding/suspend-resume.gooo"

// ExecutionSuspensionReceipt is an immutable observation of an in-flight
// execution being suspended. It does not authorize continuation or external
// access.
type ExecutionSuspensionReceipt struct {
	Source          string               `json:"source"`
	ExecutionDigest string               `json:"execution_digest"`
	SourcePhase     ExecutionPhase       `json:"source_phase"`
	Conditions      []ExecutionCondition `json:"conditions"`
	SuspendedAt     time.Time            `json:"suspended_at"`
	Reason          string               `json:"reason"`
	ReceiptDigest   string               `json:"receipt_digest"`
}

// ExecutionResumeStatus describes only the receipt observation boundary.
type ExecutionResumeStatus string

const (
	ExecutionResumeStatusResumed  ExecutionResumeStatus = "RESUMED"
	ExecutionResumeStatusRejected ExecutionResumeStatus = "REJECTED"
	ExecutionResumeStatusUnknown  ExecutionResumeStatus = "UNKNOWN"
)

// ExecutionResumeObservation records the reverse observation of a
// suspension receipt. UNKNOWN is preserved for missing or contradictory
// evidence; it is never upgraded to success by the observer.
type ExecutionResumeObservation struct {
	Source        string                `json:"source"`
	Status        ExecutionResumeStatus `json:"status"`
	ReceiptDigest string                `json:"receipt_digest"`
	ResumedAt     time.Time             `json:"resumed_at"`
	Reason        string                `json:"reason"`
}

// SuspendExecution creates a digest-bound receipt only for an explicitly
// running execution. A completed or failed execution is not suspendable.
func SuspendExecution(execution Execution, reason string, now time.Time) (ExecutionSuspensionReceipt, error) {
	if execution.Phase != ExecutionPhaseRunning {
		return ExecutionSuspensionReceipt{}, errors.New("EXECUTION_NOT_SUSPENDABLE")
	}
	if reason == "" {
		return ExecutionSuspensionReceipt{}, errors.New("EXECUTION_SUSPENSION_REASON_REQUIRED")
	}
	if now.IsZero() {
		return ExecutionSuspensionReceipt{}, errors.New("EXECUTION_SUSPENSION_TIME_REQUIRED")
	}

	receipt := ExecutionSuspensionReceipt{
		Source:          ExecutionSuspensionSource,
		ExecutionDigest: suspensionExecutionDigest(execution.Phase, execution.Conditions),
		SourcePhase:     execution.Phase,
		Conditions:      cloneExecutionConditions(execution.Conditions),
		SuspendedAt:     now,
		Reason:          reason,
	}
	receipt.ReceiptDigest = suspensionReceiptDigest(receipt)
	return receipt, nil
}

// Resume observes a suspension receipt without reconstructing or authorizing
// execution. A tampered, incomplete, or temporally impossible receipt is
// UNKNOWN rather than RESUMED.
func (receipt ExecutionSuspensionReceipt) Resume(now time.Time) ExecutionResumeObservation {
	observation := ExecutionResumeObservation{
		Source:        receipt.Source,
		ReceiptDigest: receipt.ReceiptDigest,
		ResumedAt:     now,
	}
	switch {
	case now.IsZero() || receipt.SuspendedAt.IsZero():
		observation.Status = ExecutionResumeStatusUnknown
		observation.Reason = "RESUME_TIME_UNAVAILABLE"
	case receipt.ReceiptDigest == "" || receipt.ReceiptDigest != suspensionReceiptDigest(receipt):
		observation.Status = ExecutionResumeStatusUnknown
		observation.Reason = "SUSPENSION_RECEIPT_DIGEST_MISMATCH"
	case receipt.SourcePhase != ExecutionPhaseRunning:
		observation.Status = ExecutionResumeStatusRejected
		observation.Reason = "EXECUTION_SOURCE_PHASE_NOT_RUNNING"
	case now.Before(receipt.SuspendedAt):
		observation.Status = ExecutionResumeStatusUnknown
		observation.Reason = "RESUME_BEFORE_SUSPENSION"
	default:
		observation.Status = ExecutionResumeStatusResumed
		observation.Reason = "SUSPENSION_RECEIPT_OBSERVED"
	}
	return observation
}

func cloneExecutionConditions(conditions []ExecutionCondition) []ExecutionCondition {
	if conditions == nil {
		return nil
	}
	return append([]ExecutionCondition(nil), conditions...)
}

func suspensionExecutionDigest(phase ExecutionPhase, conditions []ExecutionCondition) string {
	canonical := struct {
		Phase      ExecutionPhase       `json:"phase"`
		Conditions []ExecutionCondition `json:"conditions"`
	}{
		Phase:      phase,
		Conditions: canonicalExecutionConditions(conditions),
	}
	return suspensionJSONDigest(canonical)
}

func suspensionReceiptDigest(receipt ExecutionSuspensionReceipt) string {
	canonical := struct {
		Source          string               `json:"source"`
		ExecutionDigest string               `json:"execution_digest"`
		SourcePhase     ExecutionPhase       `json:"source_phase"`
		Conditions      []ExecutionCondition `json:"conditions"`
		SuspendedAt     time.Time            `json:"suspended_at"`
		Reason          string               `json:"reason"`
	}{
		Source:          receipt.Source,
		ExecutionDigest: receipt.ExecutionDigest,
		SourcePhase:     receipt.SourcePhase,
		Conditions:      canonicalExecutionConditions(receipt.Conditions),
		SuspendedAt:     receipt.SuspendedAt,
		Reason:          receipt.Reason,
	}
	return suspensionJSONDigest(canonical)
}

func canonicalExecutionConditions(conditions []ExecutionCondition) []ExecutionCondition {
	canonical := cloneExecutionConditions(conditions)
	sort.SliceStable(canonical, func(i, j int) bool {
		if canonical[i].Type != canonical[j].Type {
			return canonical[i].Type < canonical[j].Type
		}
		if canonical[i].Status != canonical[j].Status {
			return canonical[i].Status < canonical[j].Status
		}
		if canonical[i].Reason != canonical[j].Reason {
			return canonical[i].Reason < canonical[j].Reason
		}
		return canonical[i].Detail < canonical[j].Detail
	})
	return canonical
}

func suspensionJSONDigest(value any) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
