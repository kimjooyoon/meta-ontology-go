package bodyexecution

import (
	"context"
	"encoding/json"
)

// CompositionInputSeparation counts unique root-input tuples after source replay
// and two matching native executions. Its unit is a whole case, not an activity.
type CompositionInputSeparation struct {
	Status              string `json:"status"`
	Reason              string `json:"reason"`
	UniqueInputs        int    `json:"unique_inputs"`
	DuplicateRows       int    `json:"duplicate_rows"`
	OverlappingInputs   int    `json:"overlapping_inputs"`
	DisjointInputs      int    `json:"disjoint_inputs"`
	DisjointCasesPassed int    `json:"disjoint_cases_passed"`
	UnknownInputs       int    `json:"unknown_inputs"`
	Scope               string `json:"scope"`
}

func unknownCompositionInputSeparation(reason string) CompositionInputSeparation {
	return CompositionInputSeparation{Status: "UNKNOWN", Reason: reason,
		Scope: "unique root-input tuples; actual inputs at every assembling activity compared with source selection/holdout cases and recorded probes; all supplied expectations across duplicate rows must match; model-training exposure is unknown"}
}

type compositionInputObservation struct {
	kind   string
	passed bool
}

func (graph compositionGraph) measureInputSeparation(ctx context.Context, filename string, source []byte,
	prior Composition, suite CompositionCases, traces []CompositionTrace) CompositionInputSeparation {
	result := unknownCompositionInputSeparation("NO_DISJOINT_INPUTS")
	rows, err := graph.inputRows(suite)
	if err != nil || len(rows) != len(traces) {
		return unknownCompositionInputSeparation("INPUT_OBSERVATIONS_UNAVAILABLE")
	}
	seen, err := graph.selectionInputs(ctx, filename, source, prior)
	if err != nil {
		return unknownCompositionInputSeparation("SELECTION_INPUTS_UNAVAILABLE")
	}
	unique := make(map[string]compositionInputObservation, len(rows))
	for i, row := range rows {
		if traces[i].CaseIndex != i {
			return unknownCompositionInputSeparation("INPUT_OBSERVATIONS_UNAVAILABLE")
		}
		raw, _ := json.Marshal(row)
		key := string(raw)
		observation := graph.classifyInputTrace(seen, traces[i])
		if previous, exists := unique[key]; exists {
			observation.passed = previous.passed && observation.passed
			if previous.kind != observation.kind {
				observation.kind = "unknown"
			}
		}
		unique[key] = observation
	}
	result.UniqueInputs, result.DuplicateRows = len(unique), len(rows)-len(unique)
	for _, observation := range unique {
		switch observation.kind {
		case "overlap":
			result.OverlappingInputs++
		case "disjoint":
			result.DisjointInputs++
			if observation.passed {
				result.DisjointCasesPassed++
			}
		default:
			result.UnknownInputs++
		}
	}
	return finishInputSeparation(result)
}

func finishInputSeparation(result CompositionInputSeparation) CompositionInputSeparation {
	if result.UnknownInputs > 0 {
		result.Reason = "UNCLASSIFIED_INPUTS"
	} else if result.DisjointInputs > 0 {
		result.Status, result.Reason = "PROGRESS", "DISJOINT_EXPECTATIONS_DIFFER"
		if result.DisjointCasesPassed == result.DisjointInputs {
			result.Status, result.Reason = "PASS", "DISJOINT_EXPECTATIONS_MATCHED"
		}
	}
	return result
}

func (graph compositionGraph) classifyInputTrace(seen compositionSelectionInputs, trace CompositionTrace) compositionInputObservation {
	result := compositionInputObservation{kind: "disjoint", passed: true}
	if len(trace.Deliveries) != graph.count || !graph.hasAssembly() {
		return compositionInputObservation{kind: "unknown"}
	}
	for i, delivery := range trace.Deliveries {
		if delivery.ActivityID != graph.nodes[i].ID {
			return compositionInputObservation{kind: "unknown"}
		}
		if delivery.Passed != nil && !*delivery.Passed {
			result.passed = false
		}
		if !graph.nodes[i].Assembling {
			continue
		}
		inputs := []json.RawMessage{delivery.Input}
		if len(delivery.Inputs) > 0 {
			inputs = make([]json.RawMessage, len(delivery.Inputs))
			for p, input := range delivery.Inputs {
				inputs[p] = input.Value
			}
		}
		key, err := graph.selectionInputKey(graph.nodes[i], inputs)
		if err != nil || len(seen[i]) == 0 {
			if result.kind != "overlap" {
				result.kind = "unknown"
			}
			continue
		}
		if _, overlap := seen[i][key]; overlap {
			result.kind = "overlap"
		}
	}
	return result
}
