package selfimprovementloop

import "strings"

func Evaluate(graph Graph, in Input) (Report, error) {
	bindings, err := BindActivities(graph)
	if err != nil {
		return Report{}, err
	}
	if err := validateInput(in); err != nil {
		return Report{}, err
	}
	results, pair := evaluateCells(in, bindings)
	decisions, unknowns := summarizeResults(results)
	decision := Prioritize(decisions...)
	finalReason := "ALL_12_CELLS_CLOSED"
	var finalUnknown *UnknownState
	switch decision {
	case DecisionRefuted:
		finalReason = "REFUTED_TAKES_PRIORITY_OVER_UNKNOWN"
	case DecisionUnknown:
		finalReason = "UNKNOWN_CELL_REQUIRES_NEXT_OPERATION"
		finalUnknown = unknown(
			"PROPAGATE_OR_REFUTE", "propagate", "unresolved evidence remains",
			"INCOMPLETE_EVIDENCE", "CAPTURE_RECEIPT", strings.Join(unknownCellNames(results), ","),
		)
		unknowns = append(unknowns, *finalUnknown)
	}
	results = append(results, makeCellResult(bindingMap(bindings), "PROPAGATE_OR_REFUTE", cellEvaluation{
		Decision: decision, Reason: finalReason, Unknown: finalUnknown,
	}))
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
