package bodyexecution

import (
	"encoding/json"
	"fmt"
)

// CompositionCallInput records actual typed arguments at a constructed helper's
// function entry. RootActivityID identifies the enclosing graph invocation.
type CompositionCallInput struct {
	RootActivityID string            `json:"root_activity_id"`
	ActivityID     string            `json:"activity_id"`
	Inputs         []json.RawMessage `json:"inputs"`
}

func (graph compositionGraph) nativeCallRows(output []byte, cases int) ([][]json.RawMessage, [][]CompositionCallInput, error) {
	var rows [][]json.RawMessage
	var calls [][]CompositionCallInput
	if len(graph.plan.Preparations) == 0 {
		if err := json.Unmarshal(output, &rows); err != nil {
			return nil, nil, err
		}
	} else {
		var observed struct {
			Schema  string              `json:"schema"`
			Outputs [][]json.RawMessage `json:"outputs"`
			Calls   []struct {
				CaseIndex int `json:"case_index"`
				CompositionCallInput
			} `json:"calls"`
		}
		if err := json.Unmarshal(output, &observed); err != nil || observed.Schema != "gooo/called-input-observation/v1" || observed.Calls == nil {
			return nil, nil, fmt.Errorf("compiled composition called-input observations unavailable")
		}
		rows, calls = observed.Outputs, make([][]CompositionCallInput, cases)
		lastCase, lastRoot := -1, -1
		for _, call := range observed.Calls {
			root := graph.index(call.RootActivityID)
			node, exists := graph.called[call.ActivityID]
			if call.CaseIndex < lastCase || call.CaseIndex < 0 || call.CaseIndex >= cases || root < 0 || !exists ||
				(call.CaseIndex == lastCase && root < lastRoot) || len(calls[call.CaseIndex]) >= 65536 {
				return nil, nil, fmt.Errorf("compiled called-input identity, order or bound differs")
			}
			key, err := graph.selectionInputKey(node, call.Inputs)
			if err != nil {
				return nil, nil, err
			}
			if err := json.Unmarshal([]byte(key), &call.Inputs); err != nil {
				return nil, nil, err
			}
			calls[call.CaseIndex] = append(calls[call.CaseIndex], call.CompositionCallInput)
			lastCase, lastRoot = call.CaseIndex, root
		}
	}
	if len(rows) != cases {
		return nil, nil, fmt.Errorf("compiled composition output case count differs")
	}
	return rows, calls, nil
}
