package selfimprovementloop

import (
	"fmt"
	"sort"
	"strings"
)

type pairEvaluation struct {
	Decision string
	Reason   string
	Matched  bool
}

func Evaluate(graph Graph, in Input) (Report, error) {
	bindings, err := BindActivities(graph)
	if err != nil {
		return Report{}, err
	}
	if err := validateInput(in); err != nil {
		return Report{}, err
	}
	bindingByCell := make(map[string]ActivityBinding, len(bindings))
	for _, binding := range bindings {
		bindingByCell[binding.Cell] = binding
	}
	results := make([]CellResult, 0, len(fixedCells))
	appendResult := func(cell, decision, reason string, unknown *UnknownState) {
		binding := bindingByCell[cell]
		results = append(results, CellResult{
			Cell: cell, Activity: binding.Activity, ActivityID: binding.ActivityID,
			Decision: decision, Reason: reason, Unknown: unknown,
		})
	}

	if in.Baseline.Present && strings.TrimSpace(in.Baseline.Metric) != "" {
		appendResult("OBSERVE_BASELINE", DecisionClosed, "BASELINE_OBSERVED", nil)
	} else {
		appendResult("OBSERVE_BASELINE", DecisionUnknown, "BASELINE_NOT_OBSERVED", unknown(
			"OBSERVE_BASELINE", "observe", "baseline evidence is absent", "MISSING_EVIDENCE", "OBSERVE_BASELINE", "baseline",
		))
	}

	sameMetric := strings.TrimSpace(in.Baseline.Metric) != "" &&
		strings.TrimSpace(in.Baseline.Metric) == strings.TrimSpace(in.Target.Metric)
	if in.Target.Present && strings.TrimSpace(in.Target.Metric) != "" && sameMetric {
		appendResult("DECLARE_TARGET", DecisionClosed, "TARGET_DECLARED", nil)
	} else if in.Target.Present && strings.TrimSpace(in.Target.Metric) != "" && strings.TrimSpace(in.Baseline.Metric) != "" {
		appendResult("DECLARE_TARGET", DecisionRefuted, "TARGET_METRIC_DIFFERS_FROM_BASELINE", nil)
	} else {
		appendResult("DECLARE_TARGET", DecisionUnknown, "TARGET_NOT_DECLARED", unknown(
			"DECLARE_TARGET", "declare", "target evidence is absent", "MISSING_EVIDENCE", "DECLARE_TARGET", "target",
		))
	}

	if len(nonEmpty(in.Scope.Paths)) > 0 {
		appendResult("PIN_SCOPE", DecisionClosed, "SCOPE_PINNED", nil)
	} else {
		appendResult("PIN_SCOPE", DecisionUnknown, "SCOPE_NOT_PINNED", unknown(
			"PIN_SCOPE", "pin", "scope contains no paths", "MISSING_SCOPE", "PIN_SCOPE", "scope",
		))
	}

	appendResult("BIND_META_ACTIVITY", DecisionClosed, "RELEASED_GRAPH_ONE_TO_ONE", nil)

	if in.Transformation.Present && strings.TrimSpace(in.Transformation.Patch) != "" &&
		in.Transformation.OutputMode == "caller-owned-temporary-output" && !in.Transformation.RepositoryMutation {
		appendResult("PROPOSE_TRANSFORMATION", DecisionClosed, "PATCH_PROPOSAL_ONLY", nil)
	} else if in.Transformation.RepositoryMutation {
		appendResult("PROPOSE_TRANSFORMATION", DecisionRefuted, "REPOSITORY_MUTATION_FORBIDDEN", nil)
	} else {
		appendResult("PROPOSE_TRANSFORMATION", DecisionUnknown, "PATCH_PROPOSAL_NOT_AVAILABLE", unknown(
			"PROPOSE_TRANSFORMATION", "propose", "temporary patch proposal is absent", "MISSING_PROPOSAL", "PROPOSE_TRANSFORMATION", "transformation",
		))
	}

	predictionMatches := in.Prediction.Present && sameMetric &&
		strings.TrimSpace(in.Prediction.Metric) == strings.TrimSpace(in.Baseline.Metric) &&
		in.Prediction.Before == in.Baseline.Value && in.Prediction.After == in.Target.Value
	if predictionMatches {
		appendResult("PREDICT_EFFECT", DecisionClosed, "EFFECT_PREDICTED", nil)
	} else if in.Prediction.Present && strings.TrimSpace(in.Prediction.Metric) != "" &&
		strings.TrimSpace(in.Baseline.Metric) != "" && in.Prediction.Metric != in.Baseline.Metric {
		appendResult("PREDICT_EFFECT", DecisionRefuted, "PREDICTION_METRIC_DIFFERS", nil)
	} else {
		appendResult("PREDICT_EFFECT", DecisionUnknown, "EFFECT_NOT_PREDICTED", unknown(
			"PREDICT_EFFECT", "predict", "effect prediction is absent or unbound", "MISSING_PREDICTION", "PREDICT_EFFECT", "prediction",
		))
	}

	switch {
	case in.Counterexample.Found:
		appendResult("BUILD_COUNTEREXAMPLE", DecisionRefuted, "COUNTEREXAMPLE_FOUND", nil)
	case in.Counterexample.Checked:
		appendResult("BUILD_COUNTEREXAMPLE", DecisionClosed, "COUNTEREXAMPLE_SEARCHED", nil)
	default:
		appendResult("BUILD_COUNTEREXAMPLE", DecisionUnknown, "COUNTEREXAMPLE_NOT_BUILT", unknown(
			"BUILD_COUNTEREXAMPLE", "search", "counterexample search is absent", "MISSING_EVALUATOR", "BUILD_COUNTEREXAMPLE", "counterexample",
		))
	}

	switch {
	case in.CI.Executed && in.CI.Passed:
		appendResult("EXECUTE_CI", DecisionClosed, "CI_PASSED", nil)
	case in.CI.Executed:
		appendResult("EXECUTE_CI", DecisionRefuted, "CI_FAILED", nil)
	default:
		appendResult("EXECUTE_CI", DecisionUnknown, "CI_NOT_EXECUTED", unknown(
			"EXECUTE_CI", "execute", "CI receipt is absent", "MISSING_CI", "EXECUTE_CI", "ci",
		))
	}

	if in.Receipt.Captured && validDigest(in.Receipt.Digest) {
		appendResult("CAPTURE_RECEIPT", DecisionClosed, "RECEIPT_CAPTURED", nil)
	} else {
		appendResult("CAPTURE_RECEIPT", DecisionUnknown, "RECEIPT_NOT_CAPTURED", unknown(
			"CAPTURE_RECEIPT", "capture", "receipt digest is absent", "MISSING_RECEIPT", "CAPTURE_RECEIPT", "receipt",
		))
	}

	pair := compareExactPair(in)
	if pair.Decision == DecisionClosed {
		appendResult("COMPARE_EXACT_PAIR", pair.Decision, pair.Reason, nil)
	} else if pair.Decision == DecisionRefuted {
		appendResult("COMPARE_EXACT_PAIR", pair.Decision, pair.Reason, nil)
	} else {
		appendResult("COMPARE_EXACT_PAIR", pair.Decision, pair.Reason, unknown(
			"COMPARE_EXACT_PAIR", "match-before-after", pair.Reason, "MISSING_EXACT_INTEGER_PAIR", "CAPTURE_RECEIPT", "before_after_pair",
		))
	}

	switch strings.ToUpper(strings.TrimSpace(in.Human.Decision)) {
	case "APPROVE":
		appendResult("HUMAN_DECISION", DecisionClosed, "HUMAN_APPROVED", nil)
	case "REJECT":
		appendResult("HUMAN_DECISION", DecisionRefuted, "HUMAN_REJECTED", nil)
	default:
		appendResult("HUMAN_DECISION", DecisionUnknown, "HUMAN_DECISION_ABSENT", unknown(
			"HUMAN_DECISION", "decide", "human decision is absent", "MISSING_HUMAN_DECISION", "HUMAN_DECISION", "human",
		))
	}

	decisions := make([]string, 0, len(results))
	unknowns := make([]UnknownState, 0)
	for _, result := range results {
		decisions = append(decisions, result.Decision)
		if result.Unknown != nil {
			unknowns = append(unknowns, *result.Unknown)
		}
	}
	decision := Prioritize(decisions...)
	finalReason := "ALL_12_CELLS_CLOSED"
	var finalUnknown *UnknownState
	switch decision {
	case DecisionRefuted:
		finalReason = "REFUTED_TAKES_PRIORITY_OVER_UNKNOWN"
	case DecisionUnknown:
		finalReason = "UNKNOWN_CELL_REQUIRES_NEXT_OPERATION"
		finalUnknown = unknown(
			"PROPAGATE_OR_REFUTE", "propagate", "unresolved evidence remains", "INCOMPLETE_EVIDENCE", "CAPTURE_RECEIPT", strings.Join(unknownCellNames(results), ","),
		)
		unknowns = append(unknowns, *finalUnknown)
	}
	appendResult("PROPAGATE_OR_REFUTE", decision, finalReason, finalUnknown)

	report := Report{
		Schema: Schema, Scenario: in.Scenario, SourceDigest: in.SourceDigest,
		ToolchainDigest: in.ToolchainDigest, Decision: decision, Reason: finalReason,
		GraphHash: graph.GraphHash, Cells: results, Bindings: bindings,
		Unknowns: unknowns, PairMatched: pair.Matched,
	}
	canonical := report
	canonical.ReportDigest = ""
	report.ReportDigest = digestJSON(canonical)
	return report, nil
}

func validateInput(in Input) error {
	if in.Schema != Schema {
		return fmt.Errorf("input schema = %q, want %q", in.Schema, Schema)
	}
	if strings.TrimSpace(in.Scenario) == "" || strings.TrimSpace(in.SourceDigest) == "" || strings.TrimSpace(in.ToolchainDigest) == "" {
		return fmt.Errorf("scenario, source digest, and toolchain digest are required")
	}
	return nil
}

func compareExactPair(in Input) pairEvaluation {
	if len(in.Pair.Before) == 0 || len(in.Pair.After) == 0 {
		return pairEvaluation{Decision: DecisionUnknown, Reason: "no matching integer before/after pair"}
	}
	before := make(map[string]MetricSample, len(in.Pair.Before))
	after := make(map[string]MetricSample, len(in.Pair.After))
	for _, sample := range in.Pair.Before {
		if !sampleContextMatches(in, sample) {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "before sample context differs"}
		}
		key := sampleKey(sample)
		if _, exists := before[key]; exists {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "duplicate before sample"}
		}
		before[key] = sample
	}
	for _, sample := range in.Pair.After {
		if !sampleContextMatches(in, sample) {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "after sample context differs"}
		}
		key := sampleKey(sample)
		if _, exists := after[key]; exists {
			return pairEvaluation{Decision: DecisionRefuted, Reason: "duplicate after sample"}
		}
		after[key] = sample
	}
	keys := make([]string, 0, len(before))
	for key := range before {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, exists := after[key]; exists {
			return pairEvaluation{Decision: DecisionClosed, Reason: "EXACT_INTEGER_PAIR_MATCHED", Matched: true}
		}
	}
	return pairEvaluation{Decision: DecisionUnknown, Reason: "no matching integer before/after pair"}
}

func sampleContextMatches(in Input, sample MetricSample) bool {
	return strings.TrimSpace(sample.Scenario) == in.Scenario &&
		strings.TrimSpace(sample.SourceDigest) == in.SourceDigest &&
		strings.TrimSpace(sample.ToolchainDigest) == in.ToolchainDigest &&
		strings.TrimSpace(sample.Metric) != ""
}

func sampleKey(sample MetricSample) string {
	return sample.Scenario + "\x00" + sample.SourceDigest + "\x00" + sample.ToolchainDigest + "\x00" + sample.Metric
}

func unknown(stage, step, reason, class, next, blocked string) *UnknownState {
	return &UnknownState{
		Stage: stage, Step: step, Reason: reason, UnknownClass: class,
		NextOperation: next, BlockedBy: blocked,
	}
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func unknownCellNames(results []CellResult) []string {
	result := make([]string, 0)
	for _, cell := range results {
		if cell.Decision == DecisionUnknown {
			result = append(result, cell.Cell)
		}
	}
	return result
}
