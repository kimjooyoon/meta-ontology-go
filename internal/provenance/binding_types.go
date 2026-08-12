package provenance

import (
	"errors"
	"fmt"
	"strings"
)

const requiredCanonicalJobs = 6

// AuthoritativeEvidenceType and VerificationEvidenceType identify records that
// must carry a validated execution receipt binding.
const (
	AuthoritativeEvidenceType = "Authoritative"
	VerificationEvidenceType  = "Verification"
)

// WorkflowIdentity identifies the workflow that produced a receipt.
type WorkflowIdentity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Ref  string `json:"ref,omitempty"`
}

// JobReceipt is one of the six canonical verification jobs.
type JobReceipt struct {
	ID         string `json:"id"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
}

// CanonicalJobs contains exactly six jobs. Its wire order is normalized by ID.
type CanonicalJobs []JobReceipt

// RunBinding is the immutable CI/run tuple attached to authoritative or
// verification evidence. ReceiptBinding is an alias for adapter vocabulary.
type RunBinding struct {
	Repository      string           `json:"repository"`
	Base            string           `json:"base"`
	Head            string           `json:"head"`
	EventRef        string           `json:"event_ref"`
	CheckoutRef     string           `json:"checkout_ref"`
	RunID           int64            `json:"run_id"`
	RunAttempt      int              `json:"run_attempt"`
	Workflow        WorkflowIdentity `json:"workflow"`
	Jobs            CanonicalJobs    `json:"jobs"`
	PolicyDigest    string           `json:"policy_digest"`
	ToolchainDigest string           `json:"toolchain_digest"`
	BundleDigest    string           `json:"bundle_digest"`
	Predecessors    []string         `json:"predecessors"`
	EvidenceRefs    []string         `json:"evidence_refs"`
	WriteEffect     int              `json:"write_effect"`
}

// ReceiptBinding is the receipt-oriented name for RunBinding.
type ReceiptBinding = RunBinding

// ErrBindingRequired is returned for protected evidence without a binding.
var ErrBindingRequired = errors.New("receipt binding required")

// ErrBindingInvalid is returned when a binding tuple is incomplete or
// internally inconsistent.
var ErrBindingInvalid = errors.New("receipt binding invalid")

// ErrReplayPredecessor is returned when a predecessor is claimed twice.
var ErrReplayPredecessor = errors.New("predecessor replay")

// BindingError identifies a fail-closed binding rejection.
type BindingError struct {
	ID     string
	Kind   string
	Detail string
}

func (e *BindingError) Error() string {
	return fmt.Sprintf("evidence %q binding %s: %s", e.ID, e.Kind, e.Detail)
}

func (e *BindingError) Unwrap() error { return ErrBindingInvalid }

func (e *BindingError) Is(target error) bool {
	return target == ErrBindingInvalid || target == ErrBindingRequired && e.Kind == "missing"
}

// BindingMismatchError identifies a record that differs from an expected run
// tuple supplied by a reader.
type BindingMismatchError struct {
	ID string
}

func (e *BindingMismatchError) Error() string {
	return fmt.Sprintf("evidence %q binding does not match expected run tuple", e.ID)
}

func (e *BindingMismatchError) Unwrap() error { return ErrBindingInvalid }

// ReplayError identifies the record that tried to reuse a predecessor.
type ReplayError struct {
	ID          string
	Predecessor string
	PreviousID  string
}

func (e *ReplayError) Error() string {
	return fmt.Sprintf("evidence %q replays predecessor %q already claimed by %q", e.ID, e.Predecessor, e.PreviousID)
}

func (e *ReplayError) Unwrap() error { return ErrReplayPredecessor }

func bindingRequired(evidenceType string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(evidenceType), "_", ""), "-", ""))
	return normalized == "authoritative" || strings.HasPrefix(normalized, "verification")
}
