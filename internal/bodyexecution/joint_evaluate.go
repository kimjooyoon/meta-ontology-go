package bodyexecution

import (
	"context"
	"fmt"
)

// ConstructAndEvaluateJoint separates construction feedback from evaluation.
// A fresh construction has already executed its history; only its selected
// program is run again here. Saved receipts use ReplayJointComposition instead.
func ConstructAndEvaluateJoint(ctx context.Context, filename string, source []byte,
	construction, evaluation CompositionCases, options JointOptions) (JointConstruction, JointEvaluation, error) {
	var result JointEvaluation
	if ctx == nil {
		return JointConstruction{}, result, fmt.Errorf("joint construction requires context")
	}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, options.EntryActivity)
	if err == nil {
		_, err = graph.inputRows(evaluation)
	}
	if err != nil {
		return JointConstruction{}, result, err
	}
	prior, err := ConstructJointComposition(ctx, filename, source, construction, options)
	if err != nil {
		return prior, result, err
	}
	result.Runtime, err = ExecuteComposition(ctx, filename, []byte(prior.SelectedSource), prior.Selected, evaluation, options.GoBinary)
	if err == nil {
		result.InputSeparation, err = measureJointInputs(ctx, filename, []byte(prior.SelectedSource), prior, evaluation)
	}
	return prior, result, err
}
