package main

func buildClaimDependencyIndicators(report claimDependencyReport) []claimDependencyIndicator {
	all := claimDependencyActivities(report.Nodes, "")
	roots := claimDependencyActivities(report.Nodes, claimDependencyRootRole)
	typed := claimDependencyActivities(report.Nodes, claimDependencyEdgeRole)
	return []claimDependencyIndicator{
		{ID: "gooo.metric.claim-dependency.program-coverage.v1", Value: report.Summary.ActivitiesObserved, Target: report.Summary.ActivitiesTotal, Comparator: "EQ", Unit: "activities", Class: "DRIVER", ProofChoice: claimDependencyFoundation, Activities: all},
		{ID: "gooo.metric.claim-dependency.recoverable-roots.v1", Value: report.Summary.RecoverableRoots, Target: 1, Comparator: "GTE", Unit: "activities", Class: "DRIVER", ProofChoice: claimDependencyFoundation, Activities: roots},
		{ID: "gooo.metric.claim-dependency.typed-declarations.v1", Value: report.Summary.TypedDeclarations, Target: report.Summary.ActivitiesObserved - report.Summary.RecoverableRoots, Comparator: "EQ", Unit: "activities", Class: "DRIVER", ProofChoice: claimDependencyCoherence, Activities: typed},
		{ID: "gooo.metric.claim-dependency.edge-bindings.v1", Value: report.Summary.TypedEdges, Target: report.Summary.DependencyInputs, Comparator: "EQ", Unit: "edges", Class: "OUTCOME", ProofChoice: claimDependencyCoherence, Activities: typed},
		{ID: "gooo.metric.claim-dependency.edge-kinds.v1", Value: report.Summary.EdgeKindsObserved, Target: len(claimDependencyEdgeKinds), Comparator: "LTE", Unit: "kinds", Class: "OUTCOME", ProofChoice: claimDependencyCoherence, Activities: typed},
		{ID: "gooo.metric.claim-dependency.unresolved-inputs.v1", Value: report.Summary.UnresolvedInputs, Target: 0, Comparator: "EQ", Unit: "inputs", Class: "GUARDRAIL", ProofChoice: claimDependencyRegression, Activities: typed},
		{ID: "gooo.metric.claim-dependency.cyclic-activities.v1", Value: report.Summary.CyclicActivities, Target: 0, Comparator: "EQ", Unit: "activities", Class: "GUARDRAIL", ProofChoice: claimDependencyRegression, Activities: all},
		{ID: "gooo.metric.claim-dependency.repository-writes.v1", Value: report.Summary.RepositoryWrites, Target: 0, Comparator: "EQ", Unit: "writes", Class: "GUARDRAIL", ProofChoice: claimDependencyRegression, Activities: all},
	}
}

func claimDependencyActivities(nodes []claimDependencyNode, role string) []string {
	activities := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if role == "" || node.Role == role {
			activities = append(activities, node.Activity)
		}
	}
	return activities
}
