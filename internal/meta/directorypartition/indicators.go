package directorypartition

func summarize(applicable, roots int, targets []SourceIndicator, candidates []Candidate) Summary {
	summary := Summary{
		ApplicableIndicators: applicable, ViolatingIndicators: len(targets),
		PlannedDirectories: len(candidates), ProjectRootExemptions: roots,
	}
	for _, target := range targets {
		summary.RequiredEntries += target.Value
	}
	for _, candidate := range candidates {
		summary.PlannedEntries += len(candidate.Moves)
	}
	return summary
}

func buildIndicators(summary Summary) []Indicator {
	planCoverage := basisPoints(summary.PlannedDirectories, summary.ViolatingIndicators)
	entryCoverage := basisPoints(summary.PlannedEntries, summary.RequiredEntries)
	return []Indicator{
		{
			MetricID: "gooo.metric.meta.partition-plan.coverage-bps.v1", Class: "outcome",
			Value: planCoverage, Target: 10000, Unit: "basis_points", Relation: "greater_or_equal",
			Satisfied: planCoverage == 10000, ProofChoice: "coherence",
			MetaOperation: "resolve-directory-partition-plan", Activity: "ResolvePartitionCandidates",
		},
		{
			MetricID: "gooo.metric.meta.partition-entry-assignment.coverage-bps.v1", Class: "driver",
			Value: entryCoverage, Target: 10000, Unit: "basis_points", Relation: "greater_or_equal",
			Satisfied: entryCoverage == 10000, ProofChoice: "foundation",
			MetaOperation: "plan-directory-partitions", Activity: "PlanDirectoryPartitions",
		},
		{
			MetricID: "gooo.metric.meta.partition-repository-writes.guardrail.v1", Class: "guardrail",
			Value: summary.RepositoryWrites, Target: 0, Unit: "writes", Relation: "less_or_equal",
			Satisfied: summary.RepositoryWrites == 0, ProofChoice: "regression",
			MetaOperation: "preserve-repository-workspace", Activity: "PreserveProjectRootExemption",
		},
	}
}

func basisPoints(value, total int) int {
	if total == 0 {
		return 10000
	}
	return value * 10000 / total
}

func initialProofs(ontologyDigest, candidateDigest string) []Proof {
	return []Proof{
		{Choice: "foundation", MetaOperation: "bind-partition-ontology", Activity: "BindPartitionFoundation", Satisfied: true, EvidenceDigest: ontologyDigest},
		{Choice: "coherence", MetaOperation: "resolve-directory-partition-plan", Activity: "ResolvePartitionCandidates", Satisfied: true, EvidenceDigest: candidateDigest},
	}
}
