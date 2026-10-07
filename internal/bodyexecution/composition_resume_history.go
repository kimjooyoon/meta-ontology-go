package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// Each continuation round must use exactly the helpers constructed by that
// round. Per-record replay alone cannot establish this cross-activity relation.
func verifyCalledContinuationHistory(ctx context.Context, filename string, source []byte, prior Composition) error {
	if len(prior.Preparations) == 0 || prior.Continuation == nil {
		return nil
	}
	steps := prior.ConstructionSteps()
	rounds := -1
	for _, step := range steps {
		if r := step.Generation.Report.RecordAssembly; r != nil {
			if rounds < 0 {
				rounds = len(r.ControlHistory)
			}
			if rounds == 0 || rounds != len(r.ControlHistory) {
				return fmt.Errorf("called continuation histories do not share a construction round")
			}
		}
	}
	for stage := 0; stage < rounds; stage++ {
		current := source
		for i, step := range steps {
			if step.Generation.Report.RecordAssembly == nil {
				continue
			}
			next, err := bodycodegen.RecordAssemblyStageSource(ctx, filename, current,
				step.Generation, stage, i < len(prior.Preparations))
			if err != nil {
				return fmt.Errorf("called continuation round %d activity %s: %w", stage, step.Generation.Report.Activity, err)
			}
			current = next
		}
	}
	return nil
}
