package workspaceexecution

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Every fill here belongs to this invocation's successful construction, or to
// a replayed workspace. The measurement consumes its finite input observations.
func measureWorkspaceInputs(ctx context.Context, source []byte, result Result,
	suite bodyexecution.CompositionCases) bodyexecution.CompositionInputSeparation {
	if len(result.BodyFills) == 0 {
		return result.Runtime.InputSeparation
	}
	fills := make([]bodycodegen.Result, len(result.BodyFills))
	for i, step := range result.BodyFills {
		fills[i] = step.Generation
	}
	return bodyexecution.MeasureEarlierFillInputs(ctx, "workspace.gooo", source, result.Composition, suite, result.Runtime, fills)
}
