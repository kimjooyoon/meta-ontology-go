package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const SelfImprovementWorkspaceLifecycleObservationSchema = "gooo.self-improvement.workspace-lifecycle.v1"

const (
	SelfImprovementWorkspaceLifecycleRunning   = "RUNNING"
	SelfImprovementWorkspaceLifecycleSuspended = "SUSPENDED"
	SelfImprovementWorkspaceLifecycleResumed   = "RESUMED"
	SelfImprovementWorkspaceLifecycleUnknown   = "UNKNOWN"
)

type SelfImprovementWorkspaceLifecycleInput struct {
	TaskID          string `json:"taskId"`
	Operation       string `json:"operation"`
	WorkspaceDigest string `json:"workspaceDigest"`
	GatewayDigest   string `json:"gatewayDigest"`
	ModelDigest     string `json:"modelDigest"`
	StateDigest     string `json:"stateDigest"`
}

type SelfImprovementWorkspaceLifecycleObservation struct {
	Schema          string `json:"schema"`
	TaskID          string `json:"taskId"`
	Operation       string `json:"operation"`
	Status          string `json:"status"`
	Reason          string `json:"reason"`
	WorkspaceDigest string `json:"workspaceDigest"`
	GatewayDigest   string `json:"gatewayDigest"`
	ModelDigest     string `json:"modelDigest"`
	StateDigest     string `json:"stateDigest"`
	NonAuthorizing  bool   `json:"nonAuthorizing"`
	Digest          string `json:"digest"`
}

func ObserveSelfImprovementWorkspaceLifecycle(
	input SelfImprovementWorkspaceLifecycleInput,
) SelfImprovementWorkspaceLifecycleObservation {
	observation := SelfImprovementWorkspaceLifecycleObservation{
		Schema:          SelfImprovementWorkspaceLifecycleObservationSchema,
		TaskID:          input.TaskID,
		Operation:       input.Operation,
		WorkspaceDigest: input.WorkspaceDigest,
		GatewayDigest:   input.GatewayDigest,
		ModelDigest:     input.ModelDigest,
		StateDigest:     input.StateDigest,
		NonAuthorizing:  true,
	}

	switch {
	case input.TaskID == "":
		observation.Status = SelfImprovementWorkspaceLifecycleUnknown
		observation.Reason = "TASK_ID_MISSING"
	case input.WorkspaceDigest == "" || input.GatewayDigest == "" ||
		input.ModelDigest == "" || input.StateDigest == "":
		observation.Status = SelfImprovementWorkspaceLifecycleUnknown
		observation.Reason = "LIFECYCLE_PROVENANCE_MISSING"
	case input.Operation == "START":
		observation.Status = SelfImprovementWorkspaceLifecycleRunning
		observation.Reason = "TASK_LIFECYCLE_STARTED"
	case input.Operation == "SUSPEND":
		observation.Status = SelfImprovementWorkspaceLifecycleSuspended
		observation.Reason = "TASK_LIFECYCLE_SUSPENDED"
	case input.Operation == "RESUME":
		observation.Status = SelfImprovementWorkspaceLifecycleResumed
		observation.Reason = "TASK_LIFECYCLE_RESUMED"
	default:
		observation.Status = SelfImprovementWorkspaceLifecycleUnknown
		observation.Reason = "TASK_LIFECYCLE_OPERATION_UNKNOWN"
	}

	observation.Digest = digestSelfImprovementWorkspaceLifecycle(observation)
	return observation
}

func digestSelfImprovementWorkspaceLifecycle(
	observation SelfImprovementWorkspaceLifecycleObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}