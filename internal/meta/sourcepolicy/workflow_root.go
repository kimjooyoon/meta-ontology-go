package sourcepolicy

func workflowRootDefinition(policy Policy, observation Observation) (definition, bool) {
	if !policy.ExemptWorkflowDiscoveryRoot ||
		observation.SemanticRole != SemanticRoleWorkflowDiscoveryRoot {
		return definition{}, false
	}
	switch observation.Dimension {
	case DimensionDirectEntries, DimensionDirectoryKinds:
		return definition{family: FamilyTopology, relation: RelationObserve,
			proof: ProofFoundation, operation: OperationExemptWorkflowRoot,
			consumer: "github-actions"}, true
	default:
		return definition{}, false
	}
}
