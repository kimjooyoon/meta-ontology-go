package sourcepolicy

func candidateDefinition(operation Operation) definition {
	return definition{family: FamilyRefactor, relation: RelationEqual, blocking: false, proof: ProofRegression, operation: operation, consumer: "refactor-planner"}
}
