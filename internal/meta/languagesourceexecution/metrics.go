package languagesourceexecution

func MetricIDs() []string {
	return []string{
		"gooo.metric.language.source-execution-cases.v1",
		"gooo.metric.language.source-executions.v1",
		"gooo.metric.language.source-execution-replays.v1",
		"gooo.metric.language.runtime-diagnostic-rejections.v1",
		"gooo.metric.language.source-execution-events.v1",
		"gooo.metric.language.source-execution-unknown.guardrail.v1",
		"gooo.metric.language.source-execution-writes.guardrail.v1",
		"gooo.metric.language.source-execution-authority.guardrail.v1",
	}
}

func indicators(summary Summary) []Indicator {
	values := []struct {
		class, proof, operation string
		value, target           int
	}{
		{"OUTCOME", "FOUNDATION", "execute-source-activity", summary.CasesSatisfied, summary.CasesTotal},
		{"OUTCOME", "FOUNDATION", "execute-source-activity", summary.SourceExecutions, 1},
		{"OUTCOME", "COHERENCE", "replay-source-execution-result", summary.DeterministicReplays, 1},
		{"OUTCOME", "REGRESSION", "reject-source-runtime-failure", summary.DiagnosticRejections, 2},
		{"DRIVER", "COHERENCE", "reduce-source-execution-events", summary.ExecutionEvents, 4},
		{"GUARDRAIL", "FOUNDATION", "lower-source-execution-resolution", summary.Unknowns, 0},
		{"GUARDRAIL", "REGRESSION", "deny-source-execution-writes", summary.RepositoryWrites, 0},
		{"GUARDRAIL", "REGRESSION", "deny-source-execution-authority", summary.MutationAuthorities, 0},
	}
	result := make([]Indicator, len(values))
	for index, value := range values {
		result[index] = Indicator{MetricID: MetricIDs()[index], Class: value.class,
			ProofChoice: value.proof, MetaOperation: value.operation,
			Value: value.value, Target: value.target, Satisfied: value.value == value.target}
	}
	return result
}
