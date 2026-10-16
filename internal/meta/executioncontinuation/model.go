// Package executioncontinuation binds explicit dispatch, delivery and consumption.
package executioncontinuation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
)

const Schema = "gooo/execution-continuation/v1"
const ConsumptionSchema = "gooo/self-improvement-execution-consumption/v2"
const MaxBytes = candidatedelivery.MaxArchiveBytes

type Envelope struct {
	Repository    string `json:"source_repository"`
	WorkflowSHA   string `json:"source_head_sha"`
	SubjectSHA    string `json:"subject_sha"`
	RunID         int64  `json:"source_run_id"`
	Attempt       int    `json:"source_run_attempt"`
	ArtifactID    int64  `json:"source_artifact_id"`
	ArchiveDigest string `json:"source_artifact_digest"`
	RequestDigest string `json:"grant_request_digest"`
	Wait          string `json:"source_wait"`
	PolicyDigest  string `json:"policy_digest"`
}

type Unknown struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Class     string   `json:"unknown_class"`
	Next      string   `json:"next_operation"`
	BlockedBy []string `json:"blocked_by"`
}

type DispatchReceipt struct {
	Schema            string   `json:"schema"`
	State             string   `json:"state"`
	Reason            string   `json:"reason"`
	Source            Envelope `json:"source"`
	PolicyDigest      string   `json:"policy_digest"`
	Accepted          bool     `json:"dispatch_accepted"`
	TargetRunID       int64    `json:"target_run_id"`
	TargetWorkflowSHA string   `json:"target_workflow_sha,omitempty"`
	Unknown           *Unknown `json:"unknown,omitempty"`
	Digest            string   `json:"digest"`
}

type Metadata struct {
	Source              Envelope `json:"source"`
	ConsumerRunID       int64    `json:"consumer_run_id"`
	ConsumerAttempt     int      `json:"consumer_run_attempt"`
	ConsumerWorkflowSHA string   `json:"consumer_workflow_sha"`
	ContractRunID       int64    `json:"contract_run_id"`
	ContractAttempt     int      `json:"contract_run_attempt"`
	ContractArtifactID  int64    `json:"contract_artifact_id"`
	ContractDigest      string   `json:"contract_artifact_digest"`
	ContractWorkflowSHA string   `json:"contract_workflow_sha"`
	SubjectBlobSHA      string   `json:"subject_blob_sha"`
	SubjectFileDigest   string   `json:"subject_file_digest"`
	PolicyDigest        string   `json:"policy_digest"`
	Digest              string   `json:"digest"`
}

func Digest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }
func digestValue(v any) string { b, _ := json.Marshal(v); return Digest(b) }
func validSHA(s string) bool   { return validHex(s, 40) }
func validDigest(s string) bool {
	return strings.HasPrefix(s, "sha256:") && validHex(strings.TrimPrefix(s, "sha256:"), 64)
}
func validHex(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func (e Envelope) Validate() error {
	if !repositoryName(e.Repository) || !validSHA(e.WorkflowSHA) || !validSHA(e.SubjectSHA) || e.RunID <= 0 || e.Attempt != 1 || e.ArtifactID <= 0 || !validDigest(e.ArchiveDigest) || !validDigest(e.RequestDigest) {
		return fmt.Errorf("SOURCE_ENVELOPE_INVALID")
	}
	if e.PolicyDigest != policy.SourceDigest() {
		return fmt.Errorf("SOURCE_POLICY_MISMATCH")
	}
	if !policy.PermitWait(e.Wait, true) {
		return fmt.Errorf("COMPLETION_WAIT_CYCLE_REFUTED")
	}
	return nil
}

func (m Metadata) Validate() error {
	unsigned := m
	unsigned.Digest = ""
	if m.Digest != digestValue(unsigned) || m.PolicyDigest != policy.SourceDigest() || m.Source.Validate() != nil || m.ConsumerRunID <= 0 || m.ConsumerAttempt != 1 || !validSHA(m.ConsumerWorkflowSHA) || m.ContractRunID <= 0 || m.ContractAttempt != 1 || m.ContractArtifactID <= 0 || !validDigest(m.ContractDigest) || !validSHA(m.ContractWorkflowSHA) || !validSHA(m.SubjectBlobSHA) || !validDigest(m.SubjectFileDigest) {
		return fmt.Errorf("CONSUMER_METADATA_INVALID")
	}
	if policy.BindTargetRun(m.Source.RunID, m.ConsumerRunID, m.ConsumerRunID) != "TARGET_RUN_BOUND" {
		return fmt.Errorf("TARGET_RUN_REFUTED")
	}
	return nil
}

func sealDispatch(r DispatchReceipt, reason string) DispatchReceipt {
	r.Reason = reason
	r.Digest = ""
	if r.State == "DISPATCH_UNKNOWN" || r.State == "DISPATCH_SENT" {
		dependency := policy.MissingDependency(reason)
		r.Unknown = &Unknown{Stage: "CONTINUATION", Step: dependency, Reason: reason, Class: "DIRECT_MISSING", Next: policy.NextObservation(dependency), BlockedBy: []string{dependency}}
		if dependency == "expired_source_artifact" {
			r.Unknown.Class = "EXPIRED"
		}
	}
	r.Digest = digestValue(r)
	return r
}

func repositoryName(s string) bool {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return false
	}
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			return false
		}
		for _, c := range p {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("_.-", c)) {
				return false
			}
		}
	}
	return true
}
