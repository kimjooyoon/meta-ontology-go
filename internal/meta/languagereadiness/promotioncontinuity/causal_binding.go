package promotioncontinuity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type causalBindingProjection struct {
	HeadSHA                       string `json:"head_sha"`
	Decision                      string `json:"decision"`
	Reason                        string `json:"reason"`
	WorkspaceMode                 string `json:"workspace_mode"`
	WriteBoundary                 string `json:"write_boundary"`
	Effects                       int    `json:"effects"`
	AppliedEffects                int    `json:"applied_effects"`
	RefutedEffects                int    `json:"refuted_effects"`
	OperationOutcome              string `json:"operation_outcome"`
	ReceiptDecision               string `json:"receipt_decision"`
	ReceiptCount                  int    `json:"receipt_count"`
	FailureCount                  int    `json:"failure_count"`
	UnknownCount                  int    `json:"unknown_count"`
	DirectUnknownCount            int    `json:"direct_unknown_count"`
	DependencyBlockedUnknownCount int    `json:"dependency_blocked_unknown_count"`
	UnknownCausalDigest           string `json:"unknown_causal_digest"`
	SourceWorkspaceUnchanged      bool   `json:"source_workspace_unchanged"`
	PromotionAuthorized           bool   `json:"promotion_authorized"`
}

func causalBindingDigest(value RecoveryEvidence) string {
	projection := causalBindingProjection{
		HeadSHA: value.TransformationHeadSHA, Decision: value.TransformationDecision,
		Reason: value.TransformationReason, WorkspaceMode: value.TransformationWorkspaceMode,
		WriteBoundary: value.WriteBoundary, Effects: value.TransformationEffects,
		AppliedEffects: value.TransformationAppliedEffects, RefutedEffects: value.TransformationRefutedEffects,
		OperationOutcome: value.TransformationOperationOutcome, ReceiptDecision: value.TransformationReceiptDecision,
		ReceiptCount: value.TransformationReceiptCount, FailureCount: value.TransformationFailureCount,
		UnknownCount: value.TransformationUnknownCount, DirectUnknownCount: value.TransformationDirectUnknownCount,
		DependencyBlockedUnknownCount: value.TransformationDependencyBlockedUnknownCount,
		UnknownCausalDigest: value.TransformationUnknownCausalDigest,
		SourceWorkspaceUnchanged: value.SourceWorkspaceUnchanged, PromotionAuthorized: value.TransformationAuthorization,
	}
	payload, _ := json.Marshal(projection)
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}
