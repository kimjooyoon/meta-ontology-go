package sourcepolicy

const (
	ApplicabilityRuleDefault               = "gooo.catalog.source-policy.default-applicability.v1"
	ApplicabilityRuleProjectRootTopology   = "gooo.catalog.source-policy.project-root-topology.v1"
	ApplicabilityRuleWorkflowDiscoveryRoot = "gooo.catalog.source-policy.workflow-discovery-root.v1"
	ApplicabilityRuleProjectRootREADME     = "gooo.catalog.source-policy.project-root-readme.v1"
	ApplicabilityRuleWorkflowDiscovery     = "gooo.catalog.source-policy.github-workflow-discovery.v1"
	ApplicabilityRuleGoModuleBoundary      = "gooo.catalog.source-policy.go-module-boundary.v1"
)

const SemanticRoleWorkflowDiscoveryRoot = "workflow-discovery-root"

func isGoRewriteDimension(dimension Dimension) bool {
	switch dimension {
	case DimensionGoFileLines, DimensionFunctionLines, DimensionRefactorAssign, DimensionRefactorDuplicate:
		return true
	default:
		return false
	}
}

// ValidModuleBoundaryShape checks the original catalog meaning even when a
// module boundary prevents selecting its rewrite operation.
func ValidModuleBoundaryShape(indicator Indicator) bool {
	if !isGoRewriteDimension(indicator.MetricID) {
		return false
	}
	if indicator.Subject == "." {
		return false
	}
	observation := Observation{Subject: indicator.Subject, Dimension: indicator.MetricID}
	definition, err := definitionFor(Default(), observation)
	return err == nil && indicator.SubjectKind == indicatorSubjectKind(observation) &&
		indicator.Operation == definition.operation && indicator.Relation == definition.relation &&
		indicator.Role == definition.role && indicator.Consumer == definition.consumer &&
		indicator.Proof == definition.proof && indicator.Family == definition.family
}

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
