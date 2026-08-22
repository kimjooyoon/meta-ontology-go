package actionability

func Build(root string, in input) (Report, error) {
	authorityDigest, err := readAuthority(root)
	if err != nil {
		return Report{}, err
	}
	registry := canonicalExecutors()
	executors, err := executorIndex(registry)
	if err != nil {
		return Report{}, err
	}
	summary, operations, selected := summarize(in, executors)
	decision, reason := "FIXED_POINT", "ALL_BLOCKING_OPERATIONS_EXECUTABLE"
	if summary.UnboundIndicators != 0 {
		decision, reason = "FAIL_CLOSED", "META_BINDING_INCOMPLETE"
	} else if summary.MissingOperations != 0 {
		decision, reason = "IMPROVE", "EXECUTOR_COVERAGE_GAP"
	}
	report := Report{
		Schema: Schema, CommitSHA: in.metrics.CommitSHA, Repository: in.metrics.Repository,
		MetricsDigest: in.metricsDigest, BindingDigest: in.bindingDigest,
		RegistryDigest: digestJSON(registry), AuthorityDigest: authorityDigest,
		Decision: decision, Reason: reason, SelectedOperation: selected,
		RootProofChoice: "foundation", RootMetaOperation: "preserve-project-root-exemption",
		RootActivity: "PreserveProjectRootExemption",
		ReplayProofChoice: "regression", ReplayMetaOperation: "replay-actionability-report",
		ReplayActivity: "ReplayActionabilityReport",
		Summary: summary, Indicators: kpis(summary), Operations: operations,
	}
	report.ReportDigest = digestJSON(report)
	return report, nil
}

func kpis(summary Summary) []KPI {
	return []KPI{
		{MetricID: "gooo.metric.meta.actionable-blocking.coverage-bps.v1", Class: "outcome",
			Value: summary.IndicatorCoverageBasisPoints, Target: 10000, Unit: "basis_points",
			Relation: "greater_or_equal", Satisfied: summary.IndicatorCoverageBasisPoints >= 10000,
			ProofChoice: "coherence", Producer: "actionability.summarize", Consumer: "actionability.Build",
			MetaOperation: "resolve-indicator-executor", Activity: "ResolveIndicatorExecutor"},
		{MetricID: "gooo.metric.meta.executable-operations.coverage-bps.v1", Class: "driver",
			Value: summary.OperationCoverageBasisPoints, Target: 10000, Unit: "basis_points",
			Relation: "greater_or_equal", Satisfied: summary.OperationCoverageBasisPoints >= 10000,
			ProofChoice: "foundation", Producer: "actionability.canonicalExecutors", Consumer: "actionability.Build",
			MetaOperation: "register-executable-meta-operation", Activity: "RegisterExecutableMetaOperation"},
		{MetricID: "gooo.metric.meta.unbound-indicators.guardrail.v1", Class: "guardrail",
			Value: summary.UnboundIndicators, Target: 0, Unit: "indicators",
			Relation: "less_or_equal", Satisfied: summary.UnboundIndicators == 0,
			ProofChoice: "coherence", Producer: "metabinding.Build", Consumer: "actionability.Build",
			MetaOperation: "preserve-meta-binding-guardrail", Activity: "PreserveMetaBindingGuardrail"},
	}
}

func coverage(covered, total int) int {
	if total == 0 {
		return 10000
	}
	return covered * 10000 / total
}
