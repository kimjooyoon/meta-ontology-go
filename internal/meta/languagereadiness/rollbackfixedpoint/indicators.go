package rollbackfixedpoint

func indicators(source Source, summary Summary) []Indicator {
	recovery := recoverable(source.Guard)
	fixed := source.Transformation.Decision == "FIXED_POINT" &&
		source.Transformation.Reason == "EXACT_FIXED_POINT"
	terminal := summary.RecoveredFixedPoints + summary.AuthorizedPromotions
	return []Indicator{
		metric("rollback-fixed-point-readiness-bps", "OUTCOME", "COHERENCE", "APPLICABLE",
			summary.ReadinessBPS, 10000, summary.ReadinessBPS == 10000),
		metric("rollback-fixed-point-terminal-paths", "DRIVER", "COHERENCE", "APPLICABLE",
			terminal, 1, terminal == 1),
		metric("rollback-fixed-point-recoveries", "DRIVER", "COHERENCE", applicability(recovery),
			summary.RecoveredFixedPoints, boolInt(recovery), !recovery || fixed),
		metric("rollback-fixed-point-unresolved.guardrail", "GUARDRAIL", "FOUNDATION", "APPLICABLE",
			summary.Unresolved, 0, summary.Unresolved == 0),
		metric("rollback-fixed-point-effects.guardrail", "GUARDRAIL", "REGRESSION", applicability(recovery),
			recoveryValue(recovery, source.Transformation.Effects), 0,
			!recovery || source.Transformation.Effects == 0),
		metric("rollback-fixed-point-observer-writes.guardrail", "GUARDRAIL", "REGRESSION", "APPLICABLE",
			source.RepositoryWrites, 0, source.RepositoryWrites == 0),
		metric("rollback-fixed-point-mutation-authority.guardrail", "GUARDRAIL", "REGRESSION", "APPLICABLE",
			boolInt(source.Guard.RepositoryMutationAuthorized || source.Transformation.PromotionAuthorized),
			0, !source.Guard.RepositoryMutationAuthorized && !source.Transformation.PromotionAuthorized),
		metric("rollback-fixed-point-source-mutations.guardrail", "GUARDRAIL", "REGRESSION", "APPLICABLE",
			boolInt(!source.Transformation.SourceWorkspaceUnchanged), 0,
			source.Transformation.SourceWorkspaceUnchanged),
	}
}

func metric(id, class, choice, applicability string, value, target int, satisfied bool) Indicator {
	return Indicator{MetricID: "gooo.metric.language." + id + ".v1", Class: class,
		ProofChoice: choice, Producer: "rollbackfixedpoint.Build",
		Consumer: "self-improvement-cycle", MetaOperation: "recover-guarded-fixed-point",
		Applicability: applicability, Value: value, Target: target, Satisfied: satisfied}
}

func applicability(value bool) string {
	if value {
		return "APPLICABLE"
	}
	return "NOT_APPLICABLE"
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func recoveryValue(recovery bool, value int) int {
	if recovery {
		return value
	}
	return 0
}
