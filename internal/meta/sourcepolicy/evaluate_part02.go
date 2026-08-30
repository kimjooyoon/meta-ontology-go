package sourcepolicy

import "fmt"

func definitionFor(policy Policy, observation Observation) (definition, error) {
	observe := definition{family: FamilyVolume, relation: RelationObserve, proof: ProofCoherence, operation: OperationObserve, consumer: "metric-report"}
	if observation.Dimension == DimensionRootREADME {
		if observation.Subject != "." {
			return definition{}, fmt.Errorf("root README metric requires project-root subject")
		}
		if policy.ExemptProjectRootREADME {
			return definition{family: FamilyDocumentation, relation: RelationObserve, proof: ProofFoundation, operation: OperationExemptRootREADME, consumer: "metric-meta-program"}, nil
		}
		return definition{family: FamilyDocumentation, limit: 1, relation: RelationEqual,
			blocking: true, proof: ProofFoundation, operation: OperationRequireRootREADME,
			consumer: "repository-documenter"}, nil
	}
	if observation.Subject == ".github/workflows" && observation.Dimension == DimensionDirectEntries && observation.Detail == WorkflowDiscoveryObservationDetail {
		return definition{family: FamilyTopology, relation: RelationObserve, proof: ProofFoundation, operation: OperationPreserveWorkflow, consumer: "github-actions"}, nil
	}
	if policy.ExemptProjectRootTopology && observation.Subject == "." {
		switch observation.Dimension {
		case DimensionDirectEntries, DimensionDirectoryKinds:
			return definition{family: FamilyTopology, relation: RelationObserve,
				proof: ProofFoundation, operation: OperationExemptRoot,
				consumer: "repository-projector"}, nil
		}
	}
	if definition, ok := workflowRootDefinition(policy, observation); ok {
		return definition, nil
	}
	switch observation.Dimension {
	case DimensionGoFiles, DimensionGoooFiles, DimensionGoLines, DimensionGoooLines:
		return observe, nil
	case DimensionDirectFiles, DimensionDirectFolders, DimensionRecursiveFiles, DimensionRecursiveFolders:
		observe.family = FamilyTopology
		return observe, nil
	case DimensionGoFileLines:
		return driverCapDefinition(FamilyVolume, policy.MaxFileLines, OperationSplitGo, "source-splitter"), nil
	case DimensionGoooFileLines:
		return driverCapDefinition(FamilyVolume, policy.MaxFileLines, OperationSplitGooo, "source-splitter"), nil
	case DimensionFunctionLines:
		return driverCapDefinition(FamilyDuplication, policy.MaxFunctionLines, OperationExtractFunction, "function-extractor"), nil
	case DimensionDirectEntries:
		if policy.MaxDirectDirectoryIn == 0 {
			observe.family = FamilyTopology
			return observe, nil
		}
		return capDefinition(FamilyTopology, policy.MaxDirectDirectoryIn, OperationPartition, "repository-projector"), nil
	case DimensionDirectoryKinds:
		if !policy.RequireHomogeneousDirectories {
			observe.family = FamilyTopology
			return observe, nil
		}
		return capDefinition(FamilyTopology, 1, OperationSeparateKinds, "repository-projector"), nil
	case DimensionRefactorDuplicate:
		return capDefinition(FamilyDuplication, 0, OperationExtractFunction, "deduplicator"), nil
	case DimensionRefactorReturn:
		return definition{family: FamilyRefactor, relation: RelationObserve, blocking: false, proof: ProofCoherence, operation: OperationInspectWrapper, consumer: "refactor-report"}, nil
	case DimensionRefactorAssign:
		return candidateDefinition(OperationCollapseAssign), nil
	case DimensionFixDelta:
		return definition{family: FamilyConformance, limit: 0, relation: RelationLessOrEqual, blocking: true, proof: ProofCoherence, operation: OperationModernize, consumer: "go-fix"}, nil
	case DimensionToolchain:
		return definition{family: FamilyConformance, limit: 1, relation: RelationEqual, blocking: true, proof: ProofRegression, operation: OperationSelectToolchain, consumer: "go-toolchain-selector"}, nil
	default:
		return definition{}, fmt.Errorf("unknown source metric dimension %q", observation.Dimension)
	}
}

func capDefinition(family Family, limit int, operation Operation, consumer string) definition {
	return definition{family: family, limit: limit, relation: RelationLessOrEqual, blocking: true, proof: ProofFoundation, operation: operation, consumer: consumer}
}

func driverCapDefinition(family Family, limit int, operation Operation, consumer string) definition {
	return definition{family: family, limit: limit, relation: RelationLessOrEqual, role: IndicatorRoleDriver, proof: ProofFoundation, operation: operation, consumer: consumer}
}

func candidateDefinition(operation Operation) definition {
	return definition{family: FamilyRefactor, relation: RelationEqual, blocking: false, proof: ProofRegression, operation: operation, consumer: "refactor-planner"}
}
