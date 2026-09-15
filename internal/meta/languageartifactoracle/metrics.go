package languageartifactoracle

func MetricIDs() []string {
	return []string{
		"gooo.metric.language.artifact-oracle-cases.v1",
		"gooo.metric.language.artifact-oracle-source-bindings.v1",
		"gooo.metric.language.artifact-oracle-forgery-rejections.v1",
		"gooo.metric.language.artifact-oracle-unknown-decisions.v1",
		"gooo.metric.language.artifact-oracle-resolution-descents.v1",
		"gooo.metric.language.artifact-oracle-counterexamples.v1",
		"gooo.metric.language.artifact-oracle-producer-dependencies.guardrail.v1",
		"gooo.metric.language.artifact-oracle-unknown-checks.v1",
		"gooo.metric.language.artifact-oracle-semantic-claims.guardrail.v1",
	}
}

func indicators(summary Summary) []Indicator {
	values := []struct {
		class, proof, operation string
		value, target           int
	}{
		{"OUTCOME", "FOUNDATION", "independently-project-source", summary.CasesSatisfied, CaseTotal},
		{"OUTCOME", "FOUNDATION", "bind-artifact-to-source-bytes", summary.ExactSourceBindings, 1},
		{"GUARDRAIL", "REGRESSION", "reject-resealed-artifact-forgery", summary.ResealedForgeriesRejected, 1},
		{"GUARDRAIL", "REGRESSION", "reject-unknown-artifact-decision", summary.UnknownDecisionsRejected, 1},
		{"GUARDRAIL", "FOUNDATION", "lower-unsupported-source-resolution", summary.LowerResolutions, 1},
		{"DRIVER", "REGRESSION", "capture-shared-validator-counterexample", summary.LegacyValidatorCounterexamples, 1},
		{"GUARDRAIL", "FOUNDATION", "separate-oracle-import-graph", summary.ProducerDependencies, 0},
		{"DRIVER", "FOUNDATION", "locate-unknown-oracle-stage", summary.UnknownChecks, CheckTotal},
		{"GUARDRAIL", "FOUNDATION", "bound-oracle-semantic-claim", summary.SemanticCorrectnessClaims, 0},
	}
	result := make([]Indicator, len(values))
	for index, value := range values {
		result[index] = Indicator{MetricID: MetricIDs()[index], Class: value.class,
			ProofChoice: value.proof, MetaOperation: value.operation, Value: value.value,
			Target: value.target, Satisfied: value.value == value.target}
	}
	return result
}
