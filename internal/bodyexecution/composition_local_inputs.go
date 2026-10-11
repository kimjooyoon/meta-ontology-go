package bodyexecution

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// Consume source-replayed observations, never rerun a candidate to infer its
// historical inputs. A partial capture proves exposure for its retained prefix;
// it cannot establish that other inputs were absent during construction.
func (graph compositionGraph) includeCompositionLocalInputs(seen *compositionSelectionInputs, prior Composition) error {
	for _, step := range prior.ConstructionSteps() {
		report := step.Generation.Report
		if report.BodySearch != nil {
			if err := graph.includeLocalCallHistory(seen, report.ActivityID, report.BodySearch.LocalCalls); err != nil {
				return err
			}
		}
		if report.RecordAssembly != nil {
			for _, attempt := range report.RecordAssembly.Attempts {
				if err := graph.includeLocalCallObservation(seen, report.ActivityID, attempt.LocalCalls); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (graph compositionGraph) includeJointLocalInputs(seen *compositionSelectionInputs, attempt JointAttempt) error {
	for _, candidate := range attempt.Candidates {
		if err := graph.includeLocalCallObservation(seen, candidate.ActivityID, candidate.Attempt.LocalCalls); err != nil {
			return err
		}
	}
	for _, candidate := range attempt.SearchCandidates {
		if err := graph.includeLocalCallHistory(seen, candidate.ActivityID, candidate.LocalCalls); err != nil {
			return err
		}
	}
	return nil
}

func (graph compositionGraph) includeLocalCallHistory(seen *compositionSelectionInputs, activityID string,
	history *bodycodegen.LocalCallHistory) error {
	if history == nil {
		return nil
	}
	if history.Schema != "gooo/local-call-inputs/v1" || history.Status != "RECORDED" && history.Status != "UNKNOWN" {
		return fmt.Errorf("local input history contract differs")
	}
	if history.Status == "UNKNOWN" {
		seen.unobservedConstructionCalls = true
	}
	for _, observation := range history.Evaluations {
		if err := graph.includeLocalCallObservation(seen, activityID, observation); err != nil {
			return err
		}
	}
	return nil
}

func (graph compositionGraph) includeLocalCallObservation(seen *compositionSelectionInputs, activityID string,
	observation *bodycodegen.LocalCallObservation) error {
	if observation == nil {
		return nil // Historical absence stays absent; never clears the unknown flag.
	}
	if observation.Schema != "gooo/local-call-inputs/v1" || observation.ActivityID != activityID ||
		observation.Status != "RECORDED" && observation.Status != "UNKNOWN" {
		return fmt.Errorf("local input observation binding differs")
	}
	if observation.Status == "UNKNOWN" {
		seen.unobservedConstructionCalls = true
	}
	for _, call := range observation.Calls {
		node, set, ok := graph.constructedInputSet(call.ActivityID, seen)
		if !ok {
			continue // Fixed pure helpers have no source-owned candidate selection.
		}
		var inputs []json.RawMessage
		if err := json.Unmarshal(call.Inputs, &inputs); err != nil {
			return fmt.Errorf("local helper input tuple: %w", err)
		}
		if err := graph.includeJointTuple(node, inputs, set); err != nil {
			return err
		}
	}
	return nil
}

func (graph compositionGraph) constructedInputSet(activityID string, seen *compositionSelectionInputs) (
	CompositionActivity, map[string]struct{}, bool) {
	if node, ok := graph.called[activityID]; ok {
		return node, seen.called[activityID], true
	}
	for i, node := range graph.nodes[:graph.count] {
		if node.Assembling && node.ID == activityID {
			return node, seen.nodes[i], true
		}
	}
	return CompositionActivity{}, nil, false
}
