package directorykind

func summarize(applicable, roots int, targets []SourceIndicator, candidates []Candidate) Summary {
	summary := Summary{ApplicableIndicators: applicable, ViolatingIndicators: len(targets),
		PlannedDirectories: len(candidates), ProjectRootExemptions: roots}
	for _, candidate := range candidates {
		summary.RequiredEntries += candidate.EntryCount
		summary.PlannedEntries += len(candidate.Moves)
		summary.PlannedGroups += len(candidate.Groups)
	}
	return summary
}

func buildIndicators(summary Summary) []Indicator {
	return []Indicator{
		makeIndicator("gooo.metric.meta.directory-kind-plan.coverage-bps.v1", "outcome",
			basisPoints(summary.PlannedDirectories, summary.ViolatingIndicators), 10000, "basis_points",
			"greater_or_equal", "coherence", "resolve-directory-kind-separation", "ResolveMixedDirectoryKinds"),
		makeIndicator("gooo.metric.meta.directory-kind-entry-assignment.coverage-bps.v1", "driver",
			basisPoints(summary.PlannedEntries, summary.RequiredEntries), 10000, "basis_points",
			"greater_or_equal", "foundation", "plan-directory-kind-separation", "PlanDirectoryKindSeparation"),
		makeIndicator("gooo.metric.meta.directory-kind-group.coverage-bps.v1", "driver",
			basisPoints(summary.PlannedGroups, summary.ViolatingIndicators*2), 10000, "basis_points",
			"greater_or_equal", "coherence", "group-directory-kinds", "ResolveMixedDirectoryKinds"),
		makeIndicator("gooo.metric.meta.directory-kind-repository-writes.guardrail.v1", "guardrail",
			summary.RepositoryWrites, 0, "writes", "less_or_equal", "regression",
			"preserve-repository-workspace", "PreserveProjectRootExemption"),
	}
}

func makeIndicator(id, class string, value, target int, unit, relation, proof, operation, activity string) Indicator {
	satisfied := value >= target
	if relation == "less_or_equal" {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, Value: value, Target: target, Unit: unit,
		Relation: relation, Satisfied: satisfied, ProofChoice: proof, MetaOperation: operation, Activity: activity}
}

func basisPoints(value, total int) int {
	if total == 0 {
		return 10000
	}
	return value * 10000 / total
}
