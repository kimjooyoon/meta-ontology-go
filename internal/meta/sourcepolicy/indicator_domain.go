package sourcepolicy

const (
	ApplicabilityRuleDefault               = "gooo.catalog.source-policy.default-applicability.v1"
	ApplicabilityRuleProjectRootTopology   = "gooo.catalog.source-policy.project-root-topology.v1"
	ApplicabilityRuleWorkflowDiscoveryRoot = "gooo.catalog.source-policy.workflow-discovery-root.v1"
	ApplicabilityRuleProjectRootREADME     = "gooo.catalog.source-policy.project-root-readme.v1"
	ApplicabilityRuleWorkflowDiscovery   = "gooo.catalog.source-policy.github-workflow-discovery.v1"
)

const SemanticRoleWorkflowDiscoveryRoot = "workflow-discovery-root"

func indicatorApplicability(definition definition) (Applicability, string, ApplicabilityReason) {
	switch definition.operation {
	case OperationExemptRoot:
		return ApplicabilityNotApplicable, ApplicabilityRuleProjectRootTopology, ApplicabilityReasonRootTopologyExempt
	case OperationExemptWorkflowRoot:
		return ApplicabilityNotApplicable, ApplicabilityRuleWorkflowDiscoveryRoot, ApplicabilityReasonWorkflowRootExempt
	case OperationExemptRootREADME:
		return ApplicabilityNotApplicable, ApplicabilityRuleProjectRootREADME, ApplicabilityReasonRootREADMEExempt
	case OperationPreserveWorkflow:
		return ApplicabilityNotApplicable, ApplicabilityRuleWorkflowDiscovery, ApplicabilityReasonWorkflowDiscovery
	}
	return ApplicabilityApplicable, ApplicabilityRuleDefault, ApplicabilityReasonCatalogApplicable
}

func indicatorSubjectKind(observation Observation) SubjectKind {
	if observation.Subject == "." {
		return SubjectKindProjectRoot
	}
	switch observation.Dimension {
	case DimensionDirectFiles, DimensionDirectFolders, DimensionRecursiveFiles,
		DimensionRecursiveFolders, DimensionDirectEntries, DimensionDirectoryKinds:
		return SubjectKindDirectory
	case DimensionGoFileLines, DimensionGoooFileLines, DimensionToolchain:
		return SubjectKindFile
	case DimensionFunctionLines, DimensionRefactorReturn, DimensionRefactorAssign:
		return SubjectKindFunction
	case DimensionRefactorDuplicate:
		return SubjectKindSourceFragment
	default:
		return SubjectKindRepository
	}
}
