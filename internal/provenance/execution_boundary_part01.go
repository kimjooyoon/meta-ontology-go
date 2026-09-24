package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const ExecutionBoundaryObservationSchema = "gooo/execution-boundary-observation/v1"

type ExecutionBoundaryLifecycle string

const (
	ExecutionBoundaryRunning     ExecutionBoundaryLifecycle = "RUNNING"
	ExecutionBoundarySuspended   ExecutionBoundaryLifecycle = "SUSPENDED"
	ExecutionBoundaryFailed      ExecutionBoundaryLifecycle = "FAILED"
	ExecutionBoundaryTerminating ExecutionBoundaryLifecycle = "TERMINATING"
)

const (
	ExecutionBoundaryDecisionClosed  = "CLOSED"
	ExecutionBoundaryDecisionUnknown = "UNKNOWN"
)

type ExecutionBoundaryInput struct {
	DeclarationDigest        string
	ContractDigest           string
	IRDigest                 string
	GeneratedDigest          string
	ReverseObservationDigest string
	TaskDigest               string
	WorkspaceDigest          string
	GatewayPolicyDigest      string
	ModelDigest              string
}

type ExecutionBoundaryObservation struct {
	Schema                   string                     `json:"schema"`
	DeclarationDigest        string                     `json:"declaration_digest,omitempty"`
	ContractDigest           string                     `json:"contract_digest,omitempty"`
	IRDigest                 string                     `json:"ir_digest,omitempty"`
	GeneratedDigest          string                     `json:"generated_digest,omitempty"`
	ReverseObservationDigest string                     `json:"reverse_observation_digest,omitempty"`
	TaskDigest               string                     `json:"task_digest,omitempty"`
	WorkspaceDigest          string                     `json:"workspace_digest,omitempty"`
	GatewayPolicyDigest      string                     `json:"gateway_policy_digest,omitempty"`
	ModelDigest              string                     `json:"model_digest,omitempty"`
	Lifecycle                ExecutionBoundaryLifecycle `json:"lifecycle"`
	EvidencePrefixDigest     string                     `json:"evidence_prefix_digest"`
	MissingStageIndex        int                        `json:"missing_stage_index"`
	Decision                 string                     `json:"decision"`
	Reason                   string                     `json:"reason"`
	NonAuthorizing           bool                       `json:"non_authorizing"`
	ObservationDigest        string                     `json:"observation_digest"`
}

func ObserveExecutionBoundary(input ExecutionBoundaryInput, lifecycle ExecutionBoundaryLifecycle) ExecutionBoundaryObservation {
	value := ExecutionBoundaryObservation{
		Schema:                   ExecutionBoundaryObservationSchema,
		DeclarationDigest:        strings.TrimSpace(input.DeclarationDigest),
		ContractDigest:           strings.TrimSpace(input.ContractDigest),
		IRDigest:                 strings.TrimSpace(input.IRDigest),
		GeneratedDigest:          strings.TrimSpace(input.GeneratedDigest),
		ReverseObservationDigest: strings.TrimSpace(input.ReverseObservationDigest),
		TaskDigest:               strings.TrimSpace(input.TaskDigest),
		WorkspaceDigest:          strings.TrimSpace(input.WorkspaceDigest),
		GatewayPolicyDigest:      strings.TrimSpace(input.GatewayPolicyDigest),
		ModelDigest:              strings.TrimSpace(input.ModelDigest),
		Lifecycle:                lifecycle,
		MissingStageIndex:        -1,
		Decision:                 ExecutionBoundaryDecisionUnknown,
		Reason:                   "EXECUTION_BOUNDARY_INCOMPLETE",
		NonAuthorizing:           true,
	}
	for index, digest := range executionBoundaryStages(value) {
		if !isSHA256Digest(digest) {
			value.MissingStageIndex = index
			if strings.TrimSpace(digest) == "" {
				value.Reason = "EXECUTION_BOUNDARY_DIGEST_INCOMPLETE"
			} else {
				value.Reason = "EXECUTION_BOUNDARY_DIGEST_INVALID"
			}
			break
		}
	}
	if value.MissingStageIndex == -1 {
		if !validExecutionBoundaryLifecycle(value.Lifecycle) {
			value.Reason = "EXECUTION_BOUNDARY_LIFECYCLE_UNKNOWN"
		} else {
			value.Decision = ExecutionBoundaryDecisionClosed
			value.Reason = "EXECUTION_BOUNDARY_OBSERVED"
		}
	}
	value.EvidencePrefixDigest = executionBoundaryEvidencePrefixDigest(value)
	value.ObservationDigest = executionBoundaryObservationDigest(value)
	return value
}

func ValidateExecutionBoundaryObservation(value ExecutionBoundaryObservation) error {
	if value.Schema != ExecutionBoundaryObservationSchema || !value.NonAuthorizing || value.Reason == "" {
		return errors.New("execution boundary identity is invalid")
	}
	if value.Decision != ExecutionBoundaryDecisionClosed && value.Decision != ExecutionBoundaryDecisionUnknown {
		return errors.New("execution boundary decision is invalid")
	}
	expectedMissing := -1
	for index, digest := range executionBoundaryStages(value) {
		if digest != "" && !isSHA256Digest(digest) {
			return errors.New("execution boundary digest is invalid")
		}
		if expectedMissing == -1 && !isSHA256Digest(digest) {
			expectedMissing = index
		}
	}
	if value.MissingStageIndex != expectedMissing {
		return errors.New("execution boundary missing stage is invalid")
	}
	if value.Decision == ExecutionBoundaryDecisionClosed && (value.MissingStageIndex != -1 || !validExecutionBoundaryLifecycle(value.Lifecycle)) {
		return errors.New("closed execution boundary is incomplete")
	}
	if value.Decision == ExecutionBoundaryDecisionUnknown && value.MissingStageIndex == -1 && validExecutionBoundaryLifecycle(value.Lifecycle) {
		return errors.New("unknown execution boundary is unjustified")
	}
	if !isSHA256Digest(value.EvidencePrefixDigest) || value.EvidencePrefixDigest != executionBoundaryEvidencePrefixDigest(value) {
		return errors.New("execution boundary evidence prefix is invalid")
	}
	if !isSHA256Digest(value.ObservationDigest) || value.ObservationDigest != executionBoundaryObservationDigest(value) {
		return errors.New("execution boundary observation is invalid")
	}
	return nil
}

func executionBoundaryStages(value ExecutionBoundaryObservation) []string {
	return []string{
		value.DeclarationDigest,
		value.ContractDigest,
		value.IRDigest,
		value.GeneratedDigest,
		value.ReverseObservationDigest,
		value.TaskDigest,
		value.WorkspaceDigest,
		value.GatewayPolicyDigest,
		value.ModelDigest,
	}
}

func validExecutionBoundaryLifecycle(value ExecutionBoundaryLifecycle) bool {
	switch value {
	case ExecutionBoundaryRunning, ExecutionBoundarySuspended, ExecutionBoundaryFailed, ExecutionBoundaryTerminating:
		return true
	default:
		return false
	}
}

func executionBoundaryEvidencePrefixDigest(value ExecutionBoundaryObservation) string {
	stages := executionBoundaryStages(value)
	end := len(stages)
	if value.MissingStageIndex >= 0 && value.MissingStageIndex < end {
		end = value.MissingStageIndex
	}
	payload, _ := json.Marshal(stages[:end])
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func executionBoundaryObservationDigest(value ExecutionBoundaryObservation) string {
	value.EvidencePrefixDigest = executionBoundaryEvidencePrefixDigest(value)
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
