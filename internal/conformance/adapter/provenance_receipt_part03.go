package adapter

import (
	"fmt"
	"strings"
)

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
