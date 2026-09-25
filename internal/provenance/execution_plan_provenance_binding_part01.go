package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const ExecutionPlanProvenanceBindingSchemaPart01 = "gooo/execution-plan-provenance-binding/v1"

type ExecutionPlanLifecyclePart01 string

const (
	ExecutionPlanLifecyclePlanned   ExecutionPlanLifecyclePart01 = "PLANNED"
	ExecutionPlanLifecycleSuspended ExecutionPlanLifecyclePart01 = "SUSPENDED"
	ExecutionPlanLifecycleResumed   ExecutionPlanLifecyclePart01 = "RESUMED"
	ExecutionPlanLifecycleUnknown   ExecutionPlanLifecyclePart01 = "UNKNOWN"
)

type ExecutionPlanBindingStatusPart01 string

const (
	ExecutionPlanBindingBound   ExecutionPlanBindingStatusPart01 = "BOUND"
	ExecutionPlanBindingUnknown ExecutionPlanBindingStatusPart01 = "UNKNOWN"
)

// ExecutionPlanPart01 separates declarative task, workspace, gateway, model,
// and lifecycle identities without granting any execution or adoption power.
type ExecutionPlanPart01 struct {
	Task                          string                       `json:"task"`
	WorkspaceDigest               string                       `json:"workspace_digest"`
	GatewayPolicyDigest           string                       `json:"gateway_policy_digest"`
	Model                         string                       `json:"model"`
	Lifecycle                     ExecutionPlanLifecyclePart01 `json:"lifecycle"`
	TypedPlanDigest               string                       `json:"typed_plan_digest,omitempty"`
	ActivityOrder                 []string                     `json:"activity_order,omitempty"`
	BindingEdgeOrder              []string                     `json:"binding_edge_order,omitempty"`
	RuntimeBindingCount           int                          `json:"runtime_binding_count,omitempty"`
	WorkloadIdentityBindingDigest string                       `json:"workload_identity_binding_digest,omitempty"`
}

// ExecutionPlanProvenanceBindingPart01 binds an execution-plan observation to
// the six-stage provenance chain. It is evidence only, never authorization.
type ExecutionPlanProvenanceBindingPart01 struct {
	Schema                string                                 `json:"schema"`
	Plan                  ExecutionPlanPart01                    `json:"plan"`
	ProvenanceChainDigest string                                 `json:"provenance_chain_digest"`
	ProvenanceStages      []SelfImprovementProvenanceStagePart01 `json:"provenance_stages"`
	BoundStages           int                                    `json:"bound_stages"`
	TotalStages           int                                    `json:"total_stages"`
	NextRequiredStage     string                                 `json:"next_required_stage"`
	EvidencePrefixDigest  string                                 `json:"evidence_prefix_digest"`
	MissingStageIndex     int                                    `json:"missing_stage_index"`
	Status                ExecutionPlanBindingStatusPart01       `json:"status"`
	CausalReason          string                                 `json:"causal_reason"`
	AdoptionAuthorized    bool                                   `json:"adoption_authorized"`
	NonAuthorizing        bool                                   `json:"non_authorizing"`
	BindingDigest         string                                 `json:"binding_digest"`
}

// BindExecutionPlanToProvenancePart01 records an AX-shaped declarative
// execution boundary against the existing provenance chain. A missing or
// incomplete boundary remains UNKNOWN and never becomes execution authority.
func BindExecutionPlanToProvenancePart01(
	task,
	workspaceDigest,
	model string,
	gatewayPolicy GatewayPolicy,
	lifecycle ExecutionPlanLifecyclePart01,
	chain SelfImprovementProvenanceChainPart01,
) ExecutionPlanProvenanceBindingPart01 {
	return BindExecutionPlanToProvenanceWithTypedPlanPart01(
		task,
		workspaceDigest,
		model,
		gatewayPolicy,
		lifecycle,
		"",
		nil,
		0,
		chain,
	)
}

// BindExecutionPlanToProvenanceWithTypedPlanPart01 adds the validated
// declaration-level typed plan identity to the evidence boundary. The typed
// plan is still observation only: it does not execute activities or authorize
// adoption.
func BindExecutionPlanToProvenanceWithTypedPlanPart01(
	task,
	workspaceDigest,
	model string,
	gatewayPolicy GatewayPolicy,
	lifecycle ExecutionPlanLifecyclePart01,
	typedPlanDigest string,
	activityOrder []string,
	runtimeBindingCount int,
	chain SelfImprovementProvenanceChainPart01,
) ExecutionPlanProvenanceBindingPart01 {
	return BindExecutionPlanToProvenanceWithTypedPlanEdgesPart01(
		task,
		workspaceDigest,
		model,
		gatewayPolicy,
		lifecycle,
		typedPlanDigest,
		activityOrder,
		nil,
		runtimeBindingCount,
		chain,
	)
}

// BindExecutionPlanToProvenanceWithTypedPlanEdgesPart01 adds the validated
// declaration-level typed plan identity and canonical edge order to the
// evidence boundary. The plan remains observation only.
func BindExecutionPlanToProvenanceWithTypedPlanEdgesPart01(
	task,
	workspaceDigest,
	model string,
	gatewayPolicy GatewayPolicy,
	lifecycle ExecutionPlanLifecyclePart01,
	typedPlanDigest string,
	activityOrder []string,
	bindingEdgeOrder []string,
	runtimeBindingCount int,
	chain SelfImprovementProvenanceChainPart01,
) ExecutionPlanProvenanceBindingPart01 {
	stages := append([]SelfImprovementProvenanceStagePart01(nil), chain.Stages...)
	order := append([]string(nil), activityOrder...)
	edgeOrder := append([]string(nil), bindingEdgeOrder...)
	plan := ExecutionPlanPart01{
		Task:                strings.TrimSpace(task),
		WorkspaceDigest:     strings.TrimSpace(workspaceDigest),
		GatewayPolicyDigest: gatewayPolicy.Digest(),
		Model:               strings.TrimSpace(model),
		Lifecycle:           lifecycle,
		TypedPlanDigest:     strings.TrimSpace(typedPlanDigest),
		ActivityOrder:       order,
		BindingEdgeOrder:    edgeOrder,
		RuntimeBindingCount: runtimeBindingCount,
	}
	if !validExecutionPlanLifecyclePart01(plan.Lifecycle) {
		plan.Lifecycle = ExecutionPlanLifecycleUnknown
	}
	binding := ExecutionPlanProvenanceBindingPart01{
		Schema:                ExecutionPlanProvenanceBindingSchemaPart01,
		Plan:                  plan,
		ProvenanceChainDigest: strings.TrimSpace(chain.ChainDigest),
		ProvenanceStages:      stages,
		TotalStages:           len(stages),
		BoundStages:           countBoundExecutionPlanStagesPart01(stages),
		MissingStageIndex:     missingExecutionPlanStageIndexPart01(stages),
		Status:                ExecutionPlanBindingUnknown,
		CausalReason:          "EXECUTION_PLAN_INPUT_INCOMPLETE",
		AdoptionAuthorized:    false,
		NonAuthorizing:        true,
		NextRequiredStage:     nextExecutionPlanStagePart01(stages),
	}
	typedPlanReason := executionPlanTypedMetadataReasonPart01(plan)
	switch {
	case plan.Task == "" || plan.WorkspaceDigest == "" || plan.Model == "":
		binding.CausalReason = "EXECUTION_PLAN_INPUT_INCOMPLETE"
	case !isSHA256Digest(plan.WorkspaceDigest):
		binding.CausalReason = "EXECUTION_PLAN_WORKSPACE_DIGEST_INVALID"
	case !validExecutionPlanLifecyclePart01(lifecycle):
		binding.CausalReason = "EXECUTION_PLAN_LIFECYCLE_INVALID"
	case typedPlanReason != "":
		binding.CausalReason = typedPlanReason
	case chain.Validate() != nil:
		binding.CausalReason = "PROVENANCE_CHAIN_INVALID"
	case chain.Status != SelfImprovementProvenanceChainBoundPart01:
		binding.CausalReason = "PROVENANCE_CHAIN_UNKNOWN:" + chain.CausalReason
	default:
		binding.Status = ExecutionPlanBindingBound
		binding.CausalReason = "EXECUTION_PLAN_PROVENANCE_BOUND"
	}
	binding.EvidencePrefixDigest = executionPlanEvidencePrefixDigestPart01(binding)
	binding.BindingDigest = hashExecutionPlanProvenanceBindingPart01(binding)
	return binding
}

func BindExecutionPlanToProvenanceWithWorkloadIdentityPart01(
	task, workspaceDigest, model string,
	gatewayPolicy GatewayPolicy,
	lifecycle ExecutionPlanLifecyclePart01,
	identity WorkloadIdentityProvenanceBinding,
	chain SelfImprovementProvenanceChainPart01,
) ExecutionPlanProvenanceBindingPart01 {
	binding := BindExecutionPlanToProvenancePart01(task, workspaceDigest, model, gatewayPolicy, lifecycle, chain)
	binding.Plan.WorkloadIdentityBindingDigest = strings.TrimSpace(identity.BindingDigest)
	switch {
	case identity.Validate() != nil:
		binding.Status = ExecutionPlanBindingUnknown
		binding.CausalReason = "EXECUTION_PLAN_WORKLOAD_IDENTITY_INVALID"
	case identity.Status != WorkloadIdentityProvenanceBindingObserved:
		binding.Status = ExecutionPlanBindingUnknown
		binding.CausalReason = "EXECUTION_PLAN_WORKLOAD_IDENTITY_UNKNOWN"
	case binding.Status == ExecutionPlanBindingBound:
		binding.CausalReason = "EXECUTION_PLAN_PROVENANCE_AND_IDENTITY_BOUND"
	}
	binding.EvidencePrefixDigest = executionPlanEvidencePrefixDigestPart01(binding)
	binding.BindingDigest = hashExecutionPlanProvenanceBindingPart01(binding)
	return binding
}

func BindExecutionPlanToProvenanceWithTypedPlanAndWorkloadIdentityPart01(
	task, workspaceDigest, model string,
	gatewayPolicy GatewayPolicy,
	lifecycle ExecutionPlanLifecyclePart01,
	typedPlanDigest string,
	activityOrder []string,
	bindingEdgeOrder []string,
	runtimeBindingCount int,
	identity WorkloadIdentityProvenanceBinding,
	chain SelfImprovementProvenanceChainPart01,
) ExecutionPlanProvenanceBindingPart01 {
	binding := BindExecutionPlanToProvenanceWithTypedPlanEdgesPart01(
		task,
		workspaceDigest,
		model,
		gatewayPolicy,
		lifecycle,
		typedPlanDigest,
		activityOrder,
		bindingEdgeOrder,
		runtimeBindingCount,
		chain,
	)
	binding.Plan.WorkloadIdentityBindingDigest = strings.TrimSpace(identity.BindingDigest)
	switch {
	case identity.Validate() != nil:
		binding.Status = ExecutionPlanBindingUnknown
		binding.CausalReason = "EXECUTION_PLAN_WORKLOAD_IDENTITY_INVALID"
	case identity.Status != WorkloadIdentityProvenanceBindingObserved:
		binding.Status = ExecutionPlanBindingUnknown
		binding.CausalReason = "EXECUTION_PLAN_WORKLOAD_IDENTITY_UNKNOWN"
	case binding.Status == ExecutionPlanBindingBound:
		binding.CausalReason = "EXECUTION_PLAN_PROVENANCE_AND_IDENTITY_BOUND"
	}
	binding.EvidencePrefixDigest = executionPlanEvidencePrefixDigestPart01(binding)
	binding.BindingDigest = hashExecutionPlanProvenanceBindingPart01(binding)
	return binding
}

func (binding ExecutionPlanProvenanceBindingPart01) Validate() error {
	if binding.Schema != ExecutionPlanProvenanceBindingSchemaPart01 {
		return fmt.Errorf("unexpected execution-plan provenance schema %q", binding.Schema)
	}
	if binding.AdoptionAuthorized {
		return fmt.Errorf("execution-plan provenance binding cannot authorize adoption")
	}
	if !binding.NonAuthorizing {
		return fmt.Errorf("execution-plan provenance binding must remain non-authorizing")
	}
	if binding.TotalStages != len(binding.ProvenanceStages) {
		return fmt.Errorf("execution-plan provenance binding total stage count does not match its stages")
	}
	if reason := executionPlanTypedMetadataReasonPart01(binding.Plan); reason != "" {
		return fmt.Errorf("execution-plan typed metadata is invalid: %s", reason)
	}
	if binding.Plan.WorkloadIdentityBindingDigest != "" && !isSHA256Digest(binding.Plan.WorkloadIdentityBindingDigest) {
		return fmt.Errorf("execution-plan workload identity binding digest is invalid")
	}
	if binding.BoundStages != countBoundExecutionPlanStagesPart01(binding.ProvenanceStages) {
		return fmt.Errorf("execution-plan provenance binding bound stage count does not match its stages")
	}
	if binding.MissingStageIndex != missingExecutionPlanStageIndexPart01(binding.ProvenanceStages) {
		return fmt.Errorf("execution-plan provenance binding missing stage index does not match its stages")
	}
	if binding.NextRequiredStage != nextExecutionPlanStagePart01(binding.ProvenanceStages) {
		return fmt.Errorf("execution-plan provenance binding next stage does not match its stages")
	}
	if binding.EvidencePrefixDigest == "" ||
		binding.EvidencePrefixDigest != executionPlanEvidencePrefixDigestPart01(binding) {
		return fmt.Errorf("execution-plan provenance binding evidence prefix does not match its stages")
	}
	if binding.Status != ExecutionPlanBindingBound && binding.Status != ExecutionPlanBindingUnknown {
		return fmt.Errorf("unknown execution-plan binding status %q", binding.Status)
	}
	if binding.Status == ExecutionPlanBindingBound {
		if binding.Plan.Task == "" || binding.Plan.Model == "" {
			return fmt.Errorf("bound execution plan requires task and model identities")
		}
		if !isSHA256Digest(binding.Plan.WorkspaceDigest) {
			return fmt.Errorf("bound execution plan requires a sha256 workspace digest")
		}
		if !validExecutionPlanLifecyclePart01(binding.Plan.Lifecycle) {
			return fmt.Errorf("bound execution plan has invalid lifecycle %q", binding.Plan.Lifecycle)
		}
		if binding.ProvenanceChainDigest == "" {
			return fmt.Errorf("bound execution plan requires a provenance chain digest")
		}
		if binding.CausalReason != "EXECUTION_PLAN_PROVENANCE_BOUND" && binding.CausalReason != "EXECUTION_PLAN_PROVENANCE_AND_IDENTITY_BOUND" {
			return fmt.Errorf("bound execution plan has unexpected reason %q", binding.CausalReason)
		}
		if binding.TotalStages != 6 || binding.BoundStages != 6 ||
			binding.MissingStageIndex != -1 || binding.NextRequiredStage != "" {
			return fmt.Errorf("bound execution plan has incomplete provenance stage summary")
		}
	} else if binding.CausalReason == "" ||
		(!strings.HasPrefix(binding.CausalReason, "EXECUTION_PLAN_") && !strings.HasPrefix(binding.CausalReason, "PROVENANCE_CHAIN_")) {
		return fmt.Errorf("unknown execution plan has unexpected reason %q", binding.CausalReason)
	}
	if binding.BindingDigest == "" || binding.BindingDigest != hashExecutionPlanProvenanceBindingPart01(binding) {
		return fmt.Errorf("execution-plan provenance binding digest does not match its evidence")
	}
	return nil
}

func countBoundExecutionPlanStagesPart01(
	stages []SelfImprovementProvenanceStagePart01,
) int {
	count := 0
	for _, stage := range stages {
		if !stage.Bound {
			break
		}
		count++
	}
	return count
}

func missingExecutionPlanStageIndexPart01(
	stages []SelfImprovementProvenanceStagePart01,
) int {
	for index, stage := range stages {
		if !stage.Bound {
			return index
		}
	}
	return -1
}

func nextExecutionPlanStagePart01(
	stages []SelfImprovementProvenanceStagePart01,
) string {
	for _, stage := range stages {
		if !stage.Bound {
			return stage.Name
		}
	}
	return ""
}

func executionPlanEvidencePrefixDigestPart01(
	binding ExecutionPlanProvenanceBindingPart01,
) string {
	var canonical strings.Builder
	fmt.Fprintf(
		&canonical,
		"task=%s;workspace=%s;gateway=%s;model=%s;lifecycle=%s;typed=%s;edges=%d;chain=%s;identity=%s;status=%s;reason=%s;missing=%d;",
		binding.Plan.Task,
		binding.Plan.WorkspaceDigest,
		binding.Plan.GatewayPolicyDigest,
		binding.Plan.Model,
		binding.Plan.Lifecycle,
		binding.Plan.TypedPlanDigest,
		binding.Plan.RuntimeBindingCount,
		binding.ProvenanceChainDigest,
		binding.Plan.WorkloadIdentityBindingDigest,
		binding.Status,
		binding.CausalReason,
		binding.MissingStageIndex,
	)
	for _, activity := range binding.Plan.ActivityOrder {
		fmt.Fprintf(&canonical, "activity=%s;", activity)
	}
	for _, edge := range binding.Plan.BindingEdgeOrder {
		fmt.Fprintf(&canonical, "edge=%s;", edge)
	}
	for index, stage := range binding.ProvenanceStages {
		if binding.MissingStageIndex >= 0 && index >= binding.MissingStageIndex {
			break
		}
		fmt.Fprintf(&canonical, "%s=%s;", stage.Name, stage.Digest)
	}
	digest := sha256.Sum256([]byte(canonical.String()))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func executionPlanTypedMetadataReasonPart01(plan ExecutionPlanPart01) string {
	if plan.TypedPlanDigest == "" {
		if len(plan.ActivityOrder) != 0 || len(plan.BindingEdgeOrder) != 0 || plan.RuntimeBindingCount != 0 {
			return "EXECUTION_PLAN_TYPED_PLAN_INCOMPLETE"
		}
		return ""
	}
	if !isSHA256Digest(plan.TypedPlanDigest) {
		return "EXECUTION_PLAN_TYPED_PLAN_DIGEST_INVALID"
	}
	if len(plan.ActivityOrder) == 0 || plan.RuntimeBindingCount <= 0 {
		return "EXECUTION_PLAN_TYPED_PLAN_INCOMPLETE"
	}
	seen := make(map[string]struct{}, len(plan.ActivityOrder))
	for _, activity := range plan.ActivityOrder {
		activity = strings.TrimSpace(activity)
		if activity == "" {
			return "EXECUTION_PLAN_TYPED_PLAN_ACTIVITY_INVALID"
		}
		if _, exists := seen[activity]; exists {
			return "EXECUTION_PLAN_TYPED_PLAN_ACTIVITY_DUPLICATE"
		}
		seen[activity] = struct{}{}
	}
	if len(plan.BindingEdgeOrder) != 0 {
		if len(plan.BindingEdgeOrder) != plan.RuntimeBindingCount {
			return "EXECUTION_PLAN_TYPED_PLAN_EDGE_COUNT_MISMATCH"
		}
		seenEdges := make(map[string]struct{}, len(plan.BindingEdgeOrder))
		for _, edge := range plan.BindingEdgeOrder {
			edge = strings.TrimSpace(edge)
			if edge == "" {
				return "EXECUTION_PLAN_TYPED_PLAN_EDGE_INVALID"
			}
			if _, exists := seenEdges[edge]; exists {
				return "EXECUTION_PLAN_TYPED_PLAN_EDGE_DUPLICATE"
			}
			seenEdges[edge] = struct{}{}
		}
	}
	return ""
}

func validExecutionPlanLifecyclePart01(value ExecutionPlanLifecyclePart01) bool {
	switch value {
	case ExecutionPlanLifecyclePlanned, ExecutionPlanLifecycleSuspended, ExecutionPlanLifecycleResumed:
		return true
	default:
		return false
	}
}

func hashExecutionPlanProvenanceBindingPart01(binding ExecutionPlanProvenanceBindingPart01) string {
	binding.BindingDigest = ""
	payload, _ := json.Marshal(binding)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
