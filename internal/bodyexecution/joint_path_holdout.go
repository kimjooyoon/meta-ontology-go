package bodyexecution

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type JointPathHoldout struct {
	Evaluation     *bodycodegen.PathHoldoutReceipt `json:"evaluation"`
	CallerExposure PathHoldoutCallerExposure       `json:"caller_exposure"`
}

type PathHoldoutCallerExposure struct {
	ObservedInputs []int64 `json:"observed_inputs"`
	OtherInputs    []int64 `json:"other_inputs"`
	UnknownInputs  []int64 `json:"unknown_inputs"`
	Scope          string  `json:"scope"`
}

// Final evaluation runs after the construction decision is fixed. Its result
// records a separate observation and cannot request another selection attempt.
func (r *JointConstruction) finishSourceHoldoutEvaluation(ctx context.Context, filename string, source []byte) error {
	var err error
	r.FinalPathHoldouts, err = evaluateJointPathHoldouts(ctx, filename, source, *r)
	if err != nil {
		return err
	}
	if len(r.FinalPathHoldouts) > 0 {
		r.Schema = jointPathHoldoutSchema
		r.Scope += "; final typed-path holdouts are measured after caller selection; their scores do not alter construction decision, selection or stopping; caller input overlap remains explicit"
	}
	r.FinalRecordHoldouts, err = evaluateJointRecordHoldouts(ctx, filename, source, *r)
	if err != nil {
		return err
	}
	if len(r.FinalRecordHoldouts) > 0 {
		r.Schema = jointRecordHoldoutSchema
		r.Scope += "; final record holdouts are measured after caller selection; field and case outcomes are separate; exact caller input overlap remains explicit"
	}
	r.Stage = "COMPLETE"
	return ctx.Err()
}

// Called only after the caller's candidate loop has stopped. The earlier local
// holdout and this final selected body are separate observations; neither score
// enters betterJoint, jointComplete, a model request or a candidate's local total.
func evaluateJointPathHoldouts(ctx context.Context, filename string, source []byte,
	prior JointConstruction) ([]JointPathHoldout, error) {
	if prior.SelectedAttempt < 0 {
		return nil, nil
	}
	var results []JointPathHoldout
	for _, step := range prior.Initial.ConstructionSteps() {
		if step.Generation.Report.BodyPaths == nil {
			continue
		}
		r, err := bodycodegen.EvaluateSourcePathHoldout(ctx, filename, source, []byte(prior.SelectedSource), step.Generation.Report.Activity)
		if err != nil {
			return results, err
		}
		if r != nil {
			results = append(results, JointPathHoldout{Evaluation: r})
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
		results[i].CallerExposure = measurePathHoldoutCallerExposure(calls, &sets, known, results[i].Evaluation)
	}
	return results, nil
}

// Start with empty sets: source holdouts and local evaluation inputs must not
// become evidence that native caller feedback consumed those same inputs.
func jointHoldoutCallerInputs(calls *jointCallContext, prior JointConstruction) (compositionSelectionInputs, bool) {
	var sets compositionSelectionInputs
	if calls == nil {
		return sets, false
	}
	graph := calls.graph
	sets.called = make(map[string]map[string]struct{}, len(graph.called))
	for id := range graph.called {
		sets.called[id] = map[string]struct{}{}
	}
	for i, node := range graph.nodes[:graph.count] {
		if node.Prepared {
			sets.nodes[i] = sets.called[node.ID]
		} else if node.Assembling {
			sets.nodes[i] = map[string]struct{}{}
		}
	}
	known := true
	for _, attempt := range prior.Attempts {
		if attempt.Rejection != nil {
			continue
		}
		r := attempt.Runtime
		if !r.RuntimeReplayed || !r.ProjectionReplayed || r.CallContextSHA256 != calls.digest ||
			len(r.Traces) != len(prior.ConstructionCases.Cases) {
			known = false
			continue
		}
		for i, trace := range r.Traces {
			if trace.CaseIndex != i || graph.includeJointTrace(&sets, trace) != nil {
				known = false
			}
		}
	}
	return sets, known
}

func measurePathHoldoutCallerExposure(calls *jointCallContext, sets *compositionSelectionInputs,
	known bool, evaluation *bodycodegen.PathHoldoutReceipt) PathHoldoutCallerExposure {
	r := PathHoldoutCallerExposure{ObservedInputs: []int64{}, OtherInputs: []int64{}, UnknownInputs: []int64{},
		Scope: "exact activity arguments in every consumed native caller attempt; other means no overlap in those recorded calls; independent local helper evaluations and model-training exposure are outside this measure"}
	for _, c := range evaluation.Cases {
		if calls == nil {
			r.UnknownInputs = append(r.UnknownInputs, c.Input)
			continue
		}
		node, set, ok := calls.graph.constructedInputSet(evaluation.ActivityID, sets)
		if !ok {
			r.UnknownInputs = append(r.UnknownInputs, c.Input)
			continue
		}
		key, err := calls.graph.selectionInputKey(node, scalarInputTuple(c.Input))
		_, seen := set[key]
		switch {
		case err == nil && seen:
			r.ObservedInputs = append(r.ObservedInputs, c.Input)
		case err == nil && known:
			r.OtherInputs = append(r.OtherInputs, c.Input)
		default:
			r.UnknownInputs = append(r.UnknownInputs, c.Input)
		}
	}
	return r
}
