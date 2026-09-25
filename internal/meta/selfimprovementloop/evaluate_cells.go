package selfimprovementloop

import "strings"

type cellEvaluation struct {
	Decision string
	Reason   string
	Unknown  *UnknownState
}

func evaluateCells(in Input, bindings []ActivityBinding) ([]CellResult, pairEvaluation) {
	byCell := bindingMap(bindings)
	results := make([]CellResult, 0, len(fixedCells))
	evaluations := []cellEvaluation{
		evaluateBaseline(in), evaluateTarget(in), evaluateScope(in),
		{Decision: DecisionClosed, Reason: "RELEASED_GRAPH_ONE_TO_ONE"},
		evaluateTransformation(in), evaluatePrediction(in), evaluateCounterexample(in),
		evaluateCI(in), evaluateReceipt(in),
	}
	for index, evaluation := range evaluations {
		results = append(results, makeCellResult(byCell, fixedCells[index], evaluation))
	}
	pair := compareExactPair(in)
	results = append(results, makeCellResult(byCell, "COMPARE_EXACT_PAIR", pairCell(pair)))
	results = append(results, makeCellResult(byCell, "HUMAN_DECISION", evaluateHuman(in)))
	return results, pair
}

func evaluateBaseline(in Input) cellEvaluation {
	if in.Baseline.Present && strings.TrimSpace(in.Baseline.Metric) != "" {
		return cellEvaluation{Decision: DecisionClosed, Reason: "BASELINE_OBSERVED"}
	}
	return cellEvaluation{Decision: DecisionUnknown, Reason: "BASELINE_NOT_OBSERVED", Unknown: unknown(
		"OBSERVE_BASELINE", "observe", "baseline evidence is absent", "MISSING_EVIDENCE", "OBSERVE_BASELINE", "baseline",
	)}
}

func evaluateTarget(in Input) cellEvaluation {
	if in.Target.Present && strings.TrimSpace(in.Target.Metric) != "" && sameMetric(in) {
		return cellEvaluation{Decision: DecisionClosed, Reason: "TARGET_DECLARED"}
	}
	if in.Target.Present && strings.TrimSpace(in.Target.Metric) != "" && strings.TrimSpace(in.Baseline.Metric) != "" {
		return cellEvaluation{Decision: DecisionRefuted, Reason: "TARGET_METRIC_DIFFERS_FROM_BASELINE"}
	}
	return cellEvaluation{Decision: DecisionUnknown, Reason: "TARGET_NOT_DECLARED", Unknown: unknown(
		"DECLARE_TARGET", "declare", "target evidence is absent", "MISSING_EVIDENCE", "DECLARE_TARGET", "target",
	)}
}

func evaluateScope(in Input) cellEvaluation {
	if len(nonEmpty(in.Scope.Paths)) > 0 {
		return cellEvaluation{Decision: DecisionClosed, Reason: "SCOPE_PINNED"}
	}
	return cellEvaluation{Decision: DecisionUnknown, Reason: "SCOPE_NOT_PINNED", Unknown: unknown(
		"PIN_SCOPE", "pin", "scope contains no paths", "MISSING_SCOPE", "PIN_SCOPE", "scope",
	)}
}
