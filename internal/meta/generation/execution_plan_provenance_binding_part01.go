package generation

import "github.com/kimjooyoon/meta-ontology-go/internal/provenance"

// ExecutionPlanProvenanceBinding is the generated facade for the .gooo
// execution-plan provenance declaration. It remains evidence-only.
type ExecutionPlanProvenanceBinding = provenance.ExecutionPlanProvenanceBindingPart01

const ExecutionPlanProvenanceBindingSchemaPart01 = provenance.ExecutionPlanProvenanceBindingSchemaPart01

// BindExecutionPlanProvenancePart01 connects the declaration-level execution
// plan to its provenance chain without granting execution or adoption authority.
func BindExecutionPlanProvenancePart01(
	task,
	workspaceDigest,
	model string,
	gatewayPolicy provenance.GatewayPolicy,
	lifecycle provenance.ExecutionPlanLifecyclePart01,
	chain provenance.SelfImprovementProvenanceChainPart01,
) ExecutionPlanProvenanceBinding {
	return provenance.BindExecutionPlanToProvenancePart01(task, workspaceDigest, model, gatewayPolicy, lifecycle, chain)
}