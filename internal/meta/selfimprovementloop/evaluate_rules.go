package selfimprovementloop

import "strings"

func evaluateTransformation(in Input) cellEvaluation {
	proposal := in.Transformation
	if proposal.Present && strings.TrimSpace(proposal.Patch) != "" &&
		proposal.OutputMode == "caller-owned-temporary-output" && !proposal.RepositoryMutation {
		return cellEvaluation{Decision: DecisionClosed, Reason: "PATCH_PROPOSAL_ONLY"}
	}
	if proposal.RepositoryMutation {
		return cellEvaluation{Decision: DecisionRefuted, Reason: "REPOSITORY_MUTATION_FORBIDDEN"}
	}
	return cellEvaluation{Decision: DecisionUnknown, Reason: "PATCH_PROPOSAL_NOT_AVAILABLE", Unknown: unknown(
		"PROPOSE_TRANSFORMATION", "propose", "temporary patch proposal is absent", "MISSING_PROPOSAL", "PROPOSE_TRANSFORMATION", "transformation",
	)}
}

func evaluatePrediction(in Input) cellEvaluation {
	prediction := in.Prediction
	matches := prediction.Present && sameMetric(in) &&
		strings.TrimSpace(prediction.Metric) == strings.TrimSpace(in.Baseline.Metric) &&
		prediction.Before == in.Baseline.Value && prediction.After == in.Target.Value
	if matches {
		return cellEvaluation{Decision: DecisionClosed, Reason: "EFFECT_PREDICTED"}
	}
	if prediction.Present && strings.TrimSpace(prediction.Metric) != "" &&
		strings.TrimSpace(in.Baseline.Metric) != "" && prediction.Metric != in.Baseline.Metric {
		return cellEvaluation{Decision: DecisionRefuted, Reason: "PREDICTION_METRIC_DIFFERS"}
	}
	return cellEvaluation{Decision: DecisionUnknown, Reason: "EFFECT_NOT_PREDICTED", Unknown: unknown(
		"PREDICT_EFFECT", "predict", "effect prediction is absent or unbound", "MISSING_PREDICTION", "PREDICT_EFFECT", "prediction",
	)}
}

func evaluateCounterexample(in Input) cellEvaluation {
	switch {
	case in.Counterexample.Found:
		return cellEvaluation{Decision: DecisionRefuted, Reason: "COUNTEREXAMPLE_FOUND"}
	case in.Counterexample.Checked:
		return cellEvaluation{Decision: DecisionClosed, Reason: "COUNTEREXAMPLE_SEARCHED"}
	default:
		return cellEvaluation{Decision: DecisionUnknown, Reason: "COUNTEREXAMPLE_NOT_BUILT", Unknown: unknown(
			"BUILD_COUNTEREXAMPLE", "search", "counterexample search is absent", "MISSING_EVALUATOR", "BUILD_COUNTEREXAMPLE", "counterexample",
		)}
	}
}

func evaluateCI(in Input) cellEvaluation {
	switch {
	case in.CI.Executed && in.CI.Passed:
		return cellEvaluation{Decision: DecisionClosed, Reason: "CI_PASSED"}
	case in.CI.Executed:
		return cellEvaluation{Decision: DecisionRefuted, Reason: "CI_FAILED"}
	default:
		return cellEvaluation{Decision: DecisionUnknown, Reason: "CI_NOT_EXECUTED", Unknown: unknown(
			"EXECUTE_CI", "execute", "CI receipt is absent", "MISSING_CI", "EXECUTE_CI", "ci",
		)}
	}
}
