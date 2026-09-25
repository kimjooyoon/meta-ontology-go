package lsp

import "strings"

const executionPlanProvenanceBindingSchemaPart01 = "gooo/execution-plan-provenance-binding/v1"

type ExecutionPlanLifecycleStatePart01 string

const (
	ExecutionPlanLifecyclePlannedPart01   ExecutionPlanLifecycleStatePart01 = "PLANNED"
	ExecutionPlanLifecycleRunningPart01   ExecutionPlanLifecycleStatePart01 = "RUNNING"
	ExecutionPlanLifecycleSuspendedPart01 ExecutionPlanLifecycleStatePart01 = "SUSPENDED"
	ExecutionPlanLifecycleResumedPart01   ExecutionPlanLifecycleStatePart01 = "RESUMED"
	ExecutionPlanLifecycleCompletePart01  ExecutionPlanLifecycleStatePart01 = "COMPLETE"
	ExecutionPlanLifecycleUnknownPart01   ExecutionPlanLifecycleStatePart01 = "UNKNOWN"
	ExecutionPlanLifecycleRefutedPart01   ExecutionPlanLifecycleStatePart01 = "REFUTED"
)

type ExecutionPlanLifecycleReceiptPart01 struct {
	Schema                   string                            `json:"schema"`
	State                    ExecutionPlanLifecycleStatePart01 `json:"state"`
	StageIndex               int                               `json:"stage_index"`
	MissingStageIndex        int                               `json:"missing_stage_index"`
	EvidencePrefixDigest     string                            `json:"evidence_prefix_digest"`
	SourceDigest             string                            `json:"source_digest"`
	SemanticDigest           string                            `json:"semantic_digest"`
	TypedPlanDigest          string                            `json:"typed_plan_digest"`
	RuntimePlanDigest        string                            `json:"runtime_plan_digest"`
	GeneratedArtifactDigest  string                            `json:"generated_artifact_digest"`
	ReverseObservationDigest string                            `json:"reverse_observation_digest"`
}

func (r ExecutionPlanLifecycleReceiptPart01) FirstMissingStagePart01() int {
	if r.MissingStageIndex < 0 {
		return -1
	}
	return r.MissingStageIndex
}

func (r ExecutionPlanLifecycleReceiptPart01) CompletePart01() bool {
	return r.State == ExecutionPlanLifecycleCompletePart01 &&
		r.MissingStageIndex == -1 &&
		nonBlankExecutionPlanDigestPart01(r.EvidencePrefixDigest) &&
		nonBlankExecutionPlanDigestPart01(r.SourceDigest) &&
		nonBlankExecutionPlanDigestPart01(r.SemanticDigest) &&
		nonBlankExecutionPlanDigestPart01(r.TypedPlanDigest) &&
		nonBlankExecutionPlanDigestPart01(r.RuntimePlanDigest) &&
		nonBlankExecutionPlanDigestPart01(r.GeneratedArtifactDigest) &&
		nonBlankExecutionPlanDigestPart01(r.ReverseObservationDigest)
}

func (r ExecutionPlanLifecycleReceiptPart01) ValidPart01() bool {
	if r.Schema != executionPlanProvenanceBindingSchemaPart01 || r.StageIndex < 0 {
		return false
	}
	if !nonBlankExecutionPlanDigestPart01(r.EvidencePrefixDigest) {
		return false
	}
	if r.State == ExecutionPlanLifecycleCompletePart01 {
		return r.CompletePart01()
	}
	if r.MissingStageIndex < 0 {
		return false
	}
	switch r.State {
	case ExecutionPlanLifecyclePlannedPart01,
		ExecutionPlanLifecycleRunningPart01,
		ExecutionPlanLifecycleSuspendedPart01,
		ExecutionPlanLifecycleResumedPart01,
		ExecutionPlanLifecycleUnknownPart01,
		ExecutionPlanLifecycleRefutedPart01:
		return true
	default:
		return false
	}
}

func nonBlankExecutionPlanDigestPart01(value string) bool {
	return strings.TrimSpace(value) != ""
}