package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func evaluateJointInputs(ctx context.Context, filename string, source []byte, prior JointConstruction,
	suite CompositionCases, goBinary string, executor *Executor) (CompositionRuntime, error) {
	started := time.Now()
	calls, err := prepareJointCallContext(ctx, filename, source, prior)
	if err != nil {
		return CompositionRuntime{}, err
	}
	var result CompositionRuntime
	selected := []byte(prior.SelectedSource)
	if executor == nil {
		result, err = executeComposition(ctx, filename, selected, prior.Selected, suite, goBinary, nil, calls)
	} else {
		result, err = executor.executeWithJointCalls(ctx, filename, selected, prior.Selected, suite, goBinary, calls)
	}
	if err == nil && calls != nil {
		result.ConstructionInputHistorySHA256 = compositionDigest(prior)
		result.InputSeparation = calls.measureHistory(ctx, filename, source, prior, suite, result.Traces)
	}
	result.ElapsedNS = time.Since(started).Nanoseconds()
	return result, err
}

func (calls *jointCallContext) measureHistory(ctx context.Context, filename string, source []byte,
	prior JointConstruction, suite CompositionCases, traces []CompositionTrace) CompositionInputSeparation {
	result := unknownCompositionInputSeparation("SELECTION_INPUTS_UNAVAILABLE")
	result.Scope = "unique evaluation root-input tuples; actual constructed activity/helper arguments compared with source output/condition/holdout cases, recorded local probes and every consumed native caller attempt; duplicate expectations must all match; unobserved local calls and model-training exposure remain unknown"
	seen, err := calls.graph.selectionInputs(ctx, filename, source, prior.Initial)
	if err != nil {
		return calls.graph.unclassifiedJointInputs(suite, result)
	}
	for _, attempt := range prior.Attempts {
		if attempt.Rejection != nil {
			continue // A rejected candidate never reached the native caller.
		}
		run := attempt.Runtime
		if !run.RuntimeReplayed || !run.ProjectionReplayed || run.CallContextSHA256 != calls.digest ||
			len(run.Traces) != len(prior.ConstructionCases.Cases) {
			result.Reason = "CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED"
			return calls.graph.unclassifiedJointInputs(suite, result)
		}
		for i, trace := range run.Traces {
			if trace.CaseIndex != i || calls.graph.includeJointTrace(&seen, trace) != nil {
				result.Reason = "CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED"
				return calls.graph.unclassifiedJointInputs(suite, result)
			}
		}
	}
	result.Reason = "NO_DISJOINT_INPUTS"
	return calls.graph.measureKnownInputSeparation(suite, traces, seen, result)
}

func (graph compositionGraph) unclassifiedJointInputs(suite CompositionCases, result CompositionInputSeparation) CompositionInputSeparation {
	rows, err := graph.inputRows(suite)
	if err != nil {
		return result
	}
	unique := make(map[string]bool, len(rows))
	for _, row := range rows {
		raw, _ := json.Marshal(row)
		unique[string(raw)] = true
	}
	result.UniqueInputs, result.UnknownInputs = len(unique), len(unique)
	result.DuplicateRows = len(rows) - len(unique)
	return result
}

func (graph compositionGraph) includeJointTrace(seen *compositionSelectionInputs, trace CompositionTrace) error {
	if len(trace.Deliveries) != graph.count || len(graph.plan.Preparations) > 0 && !trace.CalledInputsObserved {
		return fmt.Errorf("joint input trace is incomplete")
	}
	for i, delivery := range trace.Deliveries {
		node := graph.nodes[i]
		if delivery.ActivityID != node.ID {
			return fmt.Errorf("joint input trace activity differs")
		}
		if !node.Assembling || len(delivery.BlockedBy) > 0 {
			continue
		}
		inputs := []json.RawMessage{delivery.Input}
		if len(delivery.Inputs) > 0 {
			inputs = make([]json.RawMessage, len(delivery.Inputs))
			for p, input := range delivery.Inputs {
				inputs[p] = input.Value
			}
		}
		if err := graph.includeJointTuple(node, inputs, seen.nodes[i]); err != nil {
			return err
		}
	}
	for _, call := range trace.Calls {
		node, ok := graph.called[call.ActivityID]
		if !ok || graph.index(call.RootActivityID) < 0 {
			return fmt.Errorf("joint input trace helper differs")
		}
		if err := graph.includeJointTuple(node, call.Inputs, seen.called[node.ID]); err != nil {
			return err
		}
	}
	return nil
}

func (graph compositionGraph) includeJointTuple(node CompositionActivity, inputs []json.RawMessage, set map[string]struct{}) error {
	if set == nil {
		return fmt.Errorf("joint input trace has no source selection set")
	}
	key, err := graph.selectionInputKey(node, inputs)
	if err == nil {
		set[key] = struct{}{}
	}
	return err
}
