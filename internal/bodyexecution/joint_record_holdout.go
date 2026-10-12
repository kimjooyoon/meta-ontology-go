package bodyexecution

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type JointRecordHoldout struct {
	Evaluation     *bodycodegen.RecordHoldoutReceipt `json:"evaluation"`
	CallerExposure RecordHoldoutCallerExposure       `json:"caller_exposure"`
}

type RecordHoldoutCallerExposure struct {
	ObservedInputs []string `json:"observed_inputs"`
	OtherInputs    []string `json:"other_inputs"`
	UnknownInputs  []string `json:"unknown_inputs"`
	Scope          string   `json:"scope"`
}

func evaluateJointRecordHoldouts(ctx context.Context, filename string, source []byte,
	prior JointConstruction) ([]JointRecordHoldout, error) {
	if prior.SelectedAttempt < 0 {
		return nil, nil
	}
	var results []JointRecordHoldout
	for _, step := range prior.Initial.ConstructionSteps() {
		if step.Generation.Report.RecordAssembly == nil {
			continue
		}
		r, err := bodycodegen.EvaluateSourceRecordHoldout(ctx, filename, source, []byte(prior.SelectedSource), step.Generation.Report.Activity)
		if err != nil {
			return results, err
		}
		if r != nil {
			results = append(results, JointRecordHoldout{Evaluation: r})
		}
	}
	if len(results) == 0 {
		return nil, nil
	}
	calls, err := prepareJointCallContext(ctx, filename, source, prior)
	if err != nil {
		return results, err
	}
	sets, known := jointHoldoutCallerInputs(calls, prior)
	for i := range results {
		results[i].CallerExposure = measureRecordHoldoutCallerExposure(calls, &sets, known, results[i].Evaluation)
	}
	return results, nil
}

func measureRecordHoldoutCallerExposure(calls *jointCallContext, sets *compositionSelectionInputs,
	known bool, evaluation *bodycodegen.RecordHoldoutReceipt) RecordHoldoutCallerExposure {
	r := RecordHoldoutCallerExposure{ObservedInputs: []string{}, OtherInputs: []string{}, UnknownInputs: []string{},
		Scope: "canonical positional input tuples in all consumed native caller attempts; other means no overlap in those recorded calls; independent local helper evaluations and model-training exposure are outside this measure"}
	for _, c := range evaluation.Cases {
		if calls == nil {
			r.UnknownInputs = append(r.UnknownInputs, c.Inputs)
			continue
		}
		node, set, ok := calls.graph.constructedInputSet(evaluation.ActivityID, sets)
		var tuple []json.RawMessage
		if !ok || json.Unmarshal([]byte(c.Inputs), &tuple) != nil {
			r.UnknownInputs = append(r.UnknownInputs, c.Inputs)
			continue
		}
		key, err := calls.graph.selectionInputKey(node, tuple)
		_, seen := set[key]
		switch {
		case err == nil && seen:
			r.ObservedInputs = append(r.ObservedInputs, c.Inputs)
		case err == nil && known:
			r.OtherInputs = append(r.OtherInputs, c.Inputs)
		default:
			r.UnknownInputs = append(r.UnknownInputs, c.Inputs)
		}
	}
	return r
}

func equalJointRecordHoldouts(a, b []JointRecordHoldout) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bodycodegen.EqualRecordHoldoutObservations(a[i].Evaluation, b[i].Evaluation) || !reflect.DeepEqual(a[i].CallerExposure, b[i].CallerExposure) {
			return false
		}
	}
	return true
}
