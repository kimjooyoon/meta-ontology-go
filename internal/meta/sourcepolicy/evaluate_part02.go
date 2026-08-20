package sourcepolicy

import "fmt"

type definition struct {
	family    Family
	limit     int
	relation  Relation
	blocking  bool
	proof     ProofChoice
	operation Operation
	consumer  string
}

func definitionFor(policy Policy, dimension Dimension) (definition, error) {
	observe := definition{family: FamilyVolume, relation: RelationObserve, proof: ProofCoherence, operation: OperationObserve, consumer: "metric-report"}
	switch dimension {
	case DimensionGoFiles, DimensionGoooFiles, DimensionGoLines, DimensionGoooLines:
		return observe, nil
	case DimensionDirectFiles, DimensionDirectFolders, DimensionRecursiveFiles, DimensionRecursiveFolders:
		observe.family = FamilyTopology
		return observe, nil
	case DimensionGoFileLines:
		return capDefinition(FamilyVolume, policy.MaxFileLines, OperationSplitGo, "source-splitter"), nil
	case DimensionGoooFileLines:
		return capDefinition(FamilyVolume, policy.MaxFileLines, OperationSplitGooo, "source-splitter"), nil
	case DimensionFunctionLines:
		return capDefinition(FamilyDuplication, policy.MaxFunctionLines, OperationExtractFunction, "function-extractor"), nil
	case DimensionDirectEntries:
		if policy.MaxDirectDirectoryIn == 0 {
			observe.family = FamilyTopology
			return observe, nil
		}
		return capDefinition(FamilyTopology, policy.MaxDirectDirectoryIn, OperationPartition, "directory-partitioner"), nil
	case DimensionDirectoryKinds:
		if !policy.RequireHomogeneousDirectories {
			observe.family = FamilyTopology
			return observe, nil
		}
		return capDefinition(FamilyTopology, 1, OperationSeparateKinds, "directory-partitioner"), nil
	case DimensionRefactorDuplicate:
		return capDefinition(FamilyDuplication, 0, OperationExtractFunction, "deduplicator"), nil
	case DimensionFixDelta:
		return definition{family: FamilyConformance, limit: 0, relation: RelationLessOrEqual, blocking: true, proof: ProofCoherence, operation: OperationModernize, consumer: "go-fix"}, nil
	case DimensionToolchain:
		return definition{family: FamilyConformance, limit: 1, relation: RelationEqual, blocking: true, proof: ProofRegression, operation: OperationSelectToolchain, consumer: "go-toolchain-selector"}, nil
	default:
		return definition{}, fmt.Errorf("unknown source metric dimension %q", dimension)
	}
}

func capDefinition(family Family, limit int, operation Operation, consumer string) definition {
	return definition{family: family, limit: limit, relation: RelationLessOrEqual, blocking: true, proof: ProofFoundation, operation: operation, consumer: consumer}
}
