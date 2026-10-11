package bodyexecution

import (
	"context"
	"fmt"
	"reflect"
)

const jointInputObservation = "gooo/joint-input-history/v1"

// This context comes from the source-bound initial construction. It only adds
// observation to the independently replayed selected projection; no saved pure
// source, preparation or plan is rewritten.
type jointCallContext struct {
	graph        compositionGraph
	preparations []CompositionStep
	digest       string
}

func prepareJointCallContext(ctx context.Context, filename string, source []byte,
	prior JointConstruction) (*jointCallContext, error) {
	if prior.InputObservation == "" {
		return nil, nil // Historical receipts did not observe these calls.
	}
	if prior.InputObservation != jointInputObservation || prior.OriginalSourceSHA256 != digest(source) {
		return nil, fmt.Errorf("joint input observation source or contract differs")
	}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, prior.Initial.Plan.EntryActivity)
	if err != nil {
		return nil, err
	}
	if !sameCompositionPlan(graph.plan, prior.Initial.Plan) {
		return nil, fmt.Errorf("joint input observation plan differs")
	}
	return &jointCallContext{graph: graph, preparations: prior.Initial.Preparations,
		digest: compositionDigest([]any{jointInputObservation, digest(source), graph.plan})}, nil
}

func (calls *jointCallContext) bind(graph compositionGraph, prior Composition) (compositionGraph, Composition, error) {
	if graph.count != calls.graph.count || !reflect.DeepEqual(graph.plan.Records, calls.graph.plan.Records) {
		return graph, prior, fmt.Errorf("joint input observation graph shape differs")
	}
	for i := range graph.count {
		old, selected := calls.graph.nodes[i], graph.nodes[i]
		old.Assembling, old.Prepared = false, false
		selected.Assembling, selected.Prepared = false, false
		if !reflect.DeepEqual(old, selected) {
			return graph, prior, fmt.Errorf("joint input observation activity shape differs")
		}
	}
	// The selected graph retains its own digest and pure projection. The copied
	// observation view supplies only the known helper identities and input types.
	graph.called = calls.graph.called
	graph.plan.Preparations = calls.graph.plan.Preparations
	prior.Preparations = calls.preparations
	return graph, prior, nil
}
