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
	Task                string                       `json:"task"`
	WorkspaceDigest     string                       `json:"workspace_digest"`
	GatewayPolicyDigest string                       `json:"gateway_policy_digest"`
	Model               string                       `json:"model"`
	Lifecycle           ExecutionPlanLifecyclePart01 `json:"lifecycle"`
}

// ExecutionPlanProvenanceBindingPart01 binds an execution-plan observation to
// the six-stage provenance chain. It is evidence only, never authorization.
type ExecutionPlanProvenanceBindingPart01 struct {
	Schema                string                           `json:"schema"`
	Plan                  ExecutionPlanPart01              `json:"plan"`
	ProvenanceChainDigest string                           `json:"provenance_chain_digest"`
	Status                ExecutionPlanBindingStatusPart01 `json:"status"`
	CausalReason          string                           `json:"causal_reason"`
	AdoptionAuthorized    bool                             `json:"adoption_authorized"`
	NonAuthorizing        bool                             `json:"non_authorizing"`
	BindingDigest         string                           `json:"binding_digest"`
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
	plan := ExecutionPlanPart01{
		Task:                strings.TrimSpace(task),
		WorkspaceDigest:     strings.TrimSpace(workspaceDigest),
		GatewayPolicyDigest: gatewayPolicy.Digest(),
		Model:               strings.TrimSpace(model),
		Lifecycle:           lifecycle,
	}
	if !validExecutionPlanLifecyclePart01(plan.Lifecycle) {
		plan.Lifecycle = ExecutionPlanLifecycleUnknown
	}
	binding := ExecutionPlanProvenanceBindingPart01{
		Schema:                ExecutionPlanProvenanceBindingSchemaPart01,
		Plan:                  plan,
		ProvenanceChainDigest: strings.TrimSpace(chain.ChainDigest),
		Status:                ExecutionPlanBindingUnknown,
		CausalReason:          "EXECUTION_PLAN_INPUT_INCOMPLETE",
		AdoptionAuthorized:    false,
		NonAuthorizing:        true,
	}
	switch {
	case plan.Task == "" || plan.WorkspaceDigest == "" || plan.Model == "":
		binding.CausalReason = "EXECUTION_PLAN_INPUT_INCOMPLETE"
	case !isSHA256Digest(plan.WorkspaceDigest):
		binding.CausalReason = "EXECUTION_PLAN_WORKSPACE_DIGEST_INVALID"
	case !validExecutionPlanLifecyclePart01(lifecycle):
		binding.CausalReason = "EXECUTION_PLAN_LIFECYCLE_INVALID"
	case chain.Validate() != nil:
		binding.CausalReason = "PROVENANCE_CHAIN_INVALID"
	case chain.Status != SelfImprovementProvenanceChainBoundPart01:
		binding.CausalReason = "PROVENANCE_CHAIN_UNKNOWN:" + chain.CausalReason
	default:
		binding.Status = ExecutionPlanBindingBound
		binding.CausalReason = "EXECUTION_PLAN_PROVENANCE_BOUND"
	}
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
		if binding.CausalReason != "EXECUTION_PLAN_PROVENANCE_BOUND" {
			return fmt.Errorf("bound execution plan has unexpected reason %q", binding.CausalReason)
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
