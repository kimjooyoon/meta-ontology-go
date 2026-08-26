package directorypartition

import "sort"

func partitionTargets(source SourceMetrics) ([]SourceIndicator, int, int) {
	targets := make([]SourceIndicator, 0)
	applicable, rootExemptions := 0, 0
	for _, indicator := range source.Meta.Indicators {
		if indicator.Subject == "." && indicator.Applicability == "NOT_APPLICABLE" {
			rootExemptions++
		}
		if indicator.MetaOperation != "partition-directory" ||
			indicator.Applicability != "APPLICABLE" || !indicator.Blocking {
			continue
		}
		applicable++
		if !indicator.Satisfied {
			targets = append(targets, indicator)
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Subject < targets[j].Subject })
	return targets, applicable, rootExemptions
}

func summarize(applicable, roots int, targets []SourceIndicator, candidates []Candidate) Summary {
	summary := Summary{
		ApplicableIndicators: applicable, ViolatingIndicators: len(targets),
		PlannedDirectories: len(candidates), ProjectRootExemptions: roots,
	}
	for _, candidate := range candidates {
		summary.RequiredEntries += candidate.EntryCount
		summary.PlannedEntries += len(candidate.Moves)
	}
	return summary
}

func buildIndicators(summary Summary) []Indicator {
	plans := basisPoints(summary.PlannedDirectories, summary.ViolatingIndicators)
	entries := basisPoints(summary.PlannedEntries, summary.RequiredEntries)
	return []Indicator{
		makeIndicator("gooo.metric.meta.partition-plan.coverage-bps.v1", "outcome", plans, 10000, "basis_points", "greater_or_equal", "coherence", "resolve-directory-partition-plan", "ResolvePartitionCandidates"),
		makeIndicator("gooo.metric.meta.partition-entry-assignment.coverage-bps.v1", "driver", entries, 10000, "basis_points", "greater_or_equal", "foundation", "plan-directory-partitions", "PlanDirectoryPartitions"),
		makeIndicator("gooo.metric.meta.partition-repository-writes.guardrail.v1", "guardrail", summary.RepositoryWrites, 0, "writes", "less_or_equal", "regression", "preserve-repository-workspace", "PreserveProjectRootExemption"),
	}
}

func makeIndicator(id, class string, value, target int, unit, relation, proof, operation, activity string) Indicator {
	satisfied := value >= target
	if relation == "less_or_equal" {
		satisfied = value <= target
	}
	return Indicator{
		MetricID: id, Class: class, Value: value, Target: target, Unit: unit,
		Relation: relation, Satisfied: satisfied, ProofChoice: proof,
		MetaOperation: operation, Activity: activity,
	}
}

func basisPoints(value, total int) int {
	if total == 0 {
		return 10000
	}
	return value * 10000 / total
}
