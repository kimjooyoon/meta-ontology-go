package proposalcompat

func indicators(source Source, summary Summary) []Indicator {
	return []Indicator{
		metric("promotion-compatibility-readiness-bps", "OUTCOME", "COHERENCE",
			summary.ReadinessBPS, 10000, summary.ReadinessBPS == 10000),
		metric("promotion-compatibility-source-v2-receipts", "DRIVER", "FOUNDATION", 1, 1,
			source.SourceSchema == "gooo/autonomous-change-proposal-promotion/v2"),
		metric("promotion-compatibility-target-v1-receipts", "DRIVER", "COHERENCE", 1, 1,
			source.TargetSchema == LegacySchema),
		metric("promotion-compatibility-projected-fields", "DRIVER", "COHERENCE",
			source.ProjectedFields, projectedFields, source.ProjectedFields == projectedFields),
		metric("promotion-compatibility-field-losses.guardrail", "GUARDRAIL", "REGRESSION",
			source.FieldLosses, 0, source.FieldLosses == 0),
		metric("promotion-compatibility-unresolved.guardrail", "GUARDRAIL", "FOUNDATION",
			summary.Unresolved, 0, summary.Unresolved == 0),
		metric("promotion-compatibility-observer-writes.guardrail", "GUARDRAIL", "REGRESSION",
			source.RepositoryWrites, 0, source.RepositoryWrites == 0),
		metric("promotion-compatibility-mutation-authority.guardrail", "GUARDRAIL", "REGRESSION",
			boolInt(source.SourceMutationAuthorized), 0, !source.SourceMutationAuthorized),
	}
}

func metric(id, class, choice string, value, target int, satisfied bool) Indicator {
	return Indicator{MetricID: "gooo.metric.language." + id + ".v1", Class: class,
		ProofChoice: choice, Producer: "proposalcompat.Build", Consumer: "guarded-promotion",
		MetaOperation: "project-promotion-contract", Value: value, Target: target,
		Satisfied: satisfied}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
