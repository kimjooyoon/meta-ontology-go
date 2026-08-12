package adapter

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ProvenanceReceiptSchema is the append-only receipt wire schema.
const ProvenanceReceiptSchema = "gooo/provenance-receipt/v1"

// ReceiptProvenanceStatus records whether the receipt is verified or deferred.
type ReceiptProvenanceStatus string

const (
	ReceiptProvenanceVerified ReceiptProvenanceStatus = "verified"
	ReceiptProvenanceDeferred ReceiptProvenanceStatus = "deferred"
)

// ReceiptOutcome describes the transaction represented by a receipt.
type ReceiptOutcome string

const (
	ReceiptOutcomeAccepted  ReceiptOutcome = "accepted"
	ReceiptOutcomeRejected  ReceiptOutcome = "rejected"
	ReceiptOutcomeCancelled ReceiptOutcome = "cancelled"
	ReceiptOutcomeClosed    ReceiptOutcome = "closed"
)

// ReceiptWriteEffect is an observer-level summary of filesystem effects.
type ReceiptWriteEffect string

const (
	ReceiptWriteEffectNone     ReceiptWriteEffect = "none"
	ReceiptWriteEffectObserved ReceiptWriteEffect = "observed"
)

// ReceiptJob is one caller-supplied check in the six-job tuple.
type ReceiptJob struct {
	Name       string `json:"name"`
	HeadSHA    string `json:"head_sha"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// ReceiptPredecessor identifies an earlier append-only evidence record.
type ReceiptPredecessor struct {
	EventRef string `json:"event_ref"`
	Digest   string `json:"digest"`
}

// ProvenanceReceipt is an append-only, deterministic verification record.
// Repository, refs, run IDs, and jobs are caller-supplied; no CI state is
// inferred or embedded by the serializer.
type ProvenanceReceipt struct {
	Schema             string                  `json:"schema"`
	Repository         string                  `json:"repository"`
	BaseSHA            string                  `json:"base_sha"`
	HeadSHA            string                  `json:"head_sha"`
	EventRef           string                  `json:"event_ref"`
	CheckoutRef        string                  `json:"checkout_ref"`
	Run                string                  `json:"run"`
	Attempt            int                     `json:"attempt"`
	ArtifactCount      int                     `json:"artifact_count"`
	Jobs               []ReceiptJob            `json:"jobs"`
	Binding            ObservationBinding      `json:"binding"`
	PreconditionDigest string                  `json:"precondition_sha256"`
	BeforeStateDigest  string                  `json:"before_state_sha256"`
	AfterStateDigest   string                  `json:"after_state_sha256"`
	Outcome            ReceiptOutcome          `json:"outcome"`
	WriteEffect        ReceiptWriteEffect      `json:"write_effect"`
	Predecessors       []ReceiptPredecessor    `json:"predecessors"`
	ProvenanceStatus   ReceiptProvenanceStatus `json:"provenance_status"`
}

var receiptJobNames = [...]string{
	"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy",
}

// CanonicalJSON serializes a validated receipt as one canonical JSONL record.
func (r ProvenanceReceipt) CanonicalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	normalized := r
	if normalized.Predecessors == nil {
		normalized.Predecessors = []ReceiptPredecessor{}
	}
	return jsonLine(normalized)
}

// Digest returns the SHA-256 digest of the canonical receipt JSONL.
func (r ProvenanceReceipt) Digest() (string, error) {
	payload, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(payload), nil
}

// ParseProvenanceReceipt accepts only canonical JSONL with known fields.
func ParseProvenanceReceipt(payload []byte) (ProvenanceReceipt, error) {
	var receipt ProvenanceReceipt
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return ProvenanceReceipt{}, fmt.Errorf("decode receipt: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return ProvenanceReceipt{}, fmt.Errorf("receipt has trailing JSON")
		}
		return ProvenanceReceipt{}, fmt.Errorf("receipt has trailing data: %w", err)
	}
	canonical, err := receipt.CanonicalJSON()
	if err != nil {
		return ProvenanceReceipt{}, err
	}
	if !bytes.Equal(payload, canonical) {
		return ProvenanceReceipt{}, fmt.Errorf("receipt is not canonical JSONL")
	}
	return receipt, nil
}

// AppendPredecessor returns a new receipt without mutating the existing record.
func (r ProvenanceReceipt) AppendPredecessor(predecessor ReceiptPredecessor) (ProvenanceReceipt, error) {
	if err := r.Validate(); err != nil {
		return ProvenanceReceipt{}, err
	}
	updated := r
	updated.Predecessors = append([]ReceiptPredecessor{}, r.Predecessors...)
	updated.Predecessors = append(updated.Predecessors, predecessor)
	if err := updated.Validate(); err != nil {
		return ProvenanceReceipt{}, err
	}
	return updated, nil
}

// Validate checks structural, digest, job, and append-order invariants.
func (r ProvenanceReceipt) Validate() error {
	if r.Schema != ProvenanceReceiptSchema {
		return fmt.Errorf("unsupported receipt schema %q", r.Schema)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "repository", value: r.Repository},
		{name: "event_ref", value: r.EventRef},
		{name: "checkout_ref", value: r.CheckoutRef},
		{name: "run", value: r.Run},
	} {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("receipt %s is required", field.name)
		}
	}
	if !validCommitID(r.BaseSHA) || !validCommitID(r.HeadSHA) {
		return fmt.Errorf("receipt base_sha and head_sha must be hexadecimal commit IDs")
	}
	if r.Attempt < 1 {
		return fmt.Errorf("receipt attempt must be positive")
	}
	if r.ArtifactCount < 0 {
		return fmt.Errorf("receipt artifact_count cannot be negative")
	}
	if r.ProvenanceStatus == ReceiptProvenanceVerified && r.ArtifactCount == 0 {
		return oracleError(OracleNW003, "verified receipt has no observer-bound artifact")
	}
	if err := validateReceiptJobs(r.Jobs, r.HeadSHA, r.ProvenanceStatus); err != nil {
		return err
	}
	if err := r.Binding.validate(); err != nil {
		return fmt.Errorf("receipt binding: %w", err)
	}
	if r.Run != r.Binding.RunID {
		return oracleError(OracleID001, "receipt run does not match observation binding")
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "precondition_sha256", value: r.PreconditionDigest},
		{name: "before_state_sha256", value: r.BeforeStateDigest},
		{name: "after_state_sha256", value: r.AfterStateDigest},
	} {
		if !validDigest(field.value) {
			return fmt.Errorf("receipt %s is not a SHA-256 digest", field.name)
		}
	}
	if err := validateReceiptOutcome(r.Outcome, r.WriteEffect); err != nil {
		return err
	}
	if r.WriteEffect == ReceiptWriteEffectNone && r.BeforeStateDigest != r.AfterStateDigest {
		return oracleError(OracleNW003, "no-write receipt has different before and after state digests")
	}
	return validateReceiptPredecessors(r.EventRef, r.Predecessors)
}

func validateReceiptOutcome(outcome ReceiptOutcome, effect ReceiptWriteEffect) error {
	switch outcome {
	case ReceiptOutcomeAccepted, ReceiptOutcomeRejected, ReceiptOutcomeCancelled, ReceiptOutcomeClosed:
	default:
		return fmt.Errorf("unsupported receipt outcome %q", outcome)
	}
	if effect != ReceiptWriteEffectNone && effect != ReceiptWriteEffectObserved {
		return fmt.Errorf("unsupported receipt write_effect %q", effect)
	}
	if outcome == ReceiptOutcomeRejected || outcome == ReceiptOutcomeCancelled || outcome == ReceiptOutcomeClosed {
		if effect != ReceiptWriteEffectNone {
			return fmt.Errorf("%s receipt must have write_effect=none", outcome)
		}
	}
	return nil
}

func validateReceiptPredecessors(eventRef string, predecessors []ReceiptPredecessor) error {
	seen := make(map[string]struct{}, len(predecessors))
	previous := ""
	for index, predecessor := range predecessors {
		if strings.TrimSpace(predecessor.EventRef) == "" || !validDigest(predecessor.Digest) {
			return fmt.Errorf("predecessor %d is incomplete", index)
		}
		if predecessor.EventRef == eventRef {
			return fmt.Errorf("receipt predecessor replays current event_ref")
		}
		if _, exists := seen[predecessor.EventRef]; exists {
			return fmt.Errorf("receipt predecessor event_ref is duplicated")
		}
		if previous != "" && predecessor.EventRef <= previous {
			return fmt.Errorf("receipt predecessors are not in canonical order")
		}
		seen[predecessor.EventRef] = struct{}{}
		previous = predecessor.EventRef
	}
	return nil
}

func validateReceiptJobs(jobs []ReceiptJob, head string, status ReceiptProvenanceStatus) error {
	if len(jobs) != len(receiptJobNames) {
		return fmt.Errorf("receipt requires exactly six jobs")
	}
	seen := make(map[string]struct{}, len(jobs))
	for _, job := range jobs {
		if !isReceiptJobName(job.Name) || job.HeadSHA != head {
			return fmt.Errorf("receipt job %q is not bound to head", job.Name)
		}
		if _, exists := seen[job.Name]; exists {
			return fmt.Errorf("receipt job %q is duplicated", job.Name)
		}
		if !validCommitID(job.HeadSHA) || !validJobStatus(job.Status) {
			return fmt.Errorf("receipt job %q is malformed", job.Name)
		}
		if job.Status == "completed" && !validJobConclusion(job.Conclusion) {
			return fmt.Errorf("receipt job %q has invalid conclusion", job.Name)
		}
		seen[job.Name] = struct{}{}
	}
	if status != ReceiptProvenanceVerified && status != ReceiptProvenanceDeferred {
		return fmt.Errorf("unsupported receipt provenance_status %q", status)
	}
	if status == ReceiptProvenanceVerified {
		for _, job := range jobs {
			if job.Status != "completed" || job.Conclusion != "success" {
				return fmt.Errorf("verified receipt has a non-success job %q", job.Name)
			}
		}
	}
	return nil
}

func isReceiptJobName(name string) bool {
	for _, expected := range receiptJobNames {
		if name == expected {
			return true
		}
	}
	return false
}

func validJobStatus(status string) bool {
	return status == "queued" || status == "in_progress" || status == "completed"
}

func validJobConclusion(conclusion string) bool {
	return conclusion == "success" || conclusion == "failure" ||
		conclusion == "cancelled" || conclusion == "skipped" || conclusion == "neutral"
}

func validCommitID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
