package generation

func actionMatchesSourceIndicator(action Action) bool {
	indicator := action.SourceIndicator
	return indicatorID(indicator) == action.IndicatorID &&
		!indicator.Satisfied &&
		indicator.MetricID == action.MetricID &&
		indicator.Subject == action.Subject &&
		indicator.SubjectKind == action.SubjectKind &&
		indicator.Applicability == action.Applicability &&
		indicator.ApplicabilityRule == action.ApplicabilityRule &&
		indicator.ApplicabilityReason == action.ApplicabilityReason &&
		indicator.Blocking == action.Blocking &&
		indicator.Proof == action.MetricProofChoice &&
		indicator.Producer == action.MetricProducer &&
		indicator.Consumer == action.MetricConsumer &&
		indicator.Operation == action.Operation
}

func stepMatchesSourceIndicator(step ExecutionStep) bool {
	indicator := step.SourceIndicator
	return indicatorID(indicator) == step.ActionIndicatorID &&
		!indicator.Satisfied &&
		indicator.MetricID == step.MetricID &&
		indicator.Subject == step.Subject &&
		indicator.SubjectKind == step.SubjectKind &&
		indicator.Applicability == step.Applicability &&
		indicator.ApplicabilityRule == step.ApplicabilityRule &&
		indicator.ApplicabilityReason == step.ApplicabilityReason &&
		indicator.Blocking == step.Blocking &&
		indicator.Proof == step.MetricProofChoice &&
		indicator.Producer == step.MetricProducer &&
		indicator.Consumer == step.MetricConsumer &&
		indicator.Operation == step.Operation
}
