package sourcepolicy

const (
	ApplicabilityRuleDefault             = "gooo.catalog.source-policy.default-applicability.v1"
	ApplicabilityRuleProjectRootTopology = "gooo.catalog.source-policy.project-root-topology.v1"
)

func indicatorApplicability(definition definition) (Applicability, string, ApplicabilityReason) {
	if definition.operation == OperationExemptRoot {
		return ApplicabilityNotApplicable, ApplicabilityRuleProjectRootTopology, ApplicabilityReasonRootTopologyExempt
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
