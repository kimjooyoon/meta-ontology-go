package bodyrefinement

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func reviseRound(ctx context.Context, filename string, current []byte, options Options, round *Round) ([]byte, error) {
	var err error
	if round.Decision.Action == "INCORPORATE" {
		current, round.FeedbackAdded, err = bodycodegen.ExtendAssemblyCases(ctx, filename, current, options.Activity, counterexamples(*round, options.Activity))
		if err != nil {
			return nil, err
		}
		round.RevisedSource = string(current)
		if round.FeedbackAdded == 0 && round.Decision.NextAttempts == round.Observation.Attempts {
			return nil, fmt.Errorf("feedback incorporation made no source or budget progress")
		}
	}
	current, err = bodycodegen.ReviseAssemblyBudget(ctx, filename, current, options.Activity, round.Decision.NextAttempts)
	if err == nil {
		round.RevisedSource = string(current)
	}
	return current, err
}

// Only direct caller inputs supply new obligations. An upstream program's
// incorrect output must not silently become a downstream training input.
func counterexamples(round Round, activity string) []bodycodegen.AssemblyFeedback {
	id := ""
	for _, node := range round.Composition.Plan.Activities {
		if node.Name == activity {
			id = node.ID
		}
	}
	var rows []bodycodegen.AssemblyFeedback
	for _, trace := range round.Runtime.Traces {
		for _, value := range trace.Deliveries {
			if value.ActivityID != id || value.Passed == nil || *value.Passed || value.ProducerID != "" {
				continue
			}
			inputs := []json.RawMessage{value.Input}
			if len(value.Inputs) > 0 {
				inputs = nil
				for _, port := range value.Inputs {
					if port.ProducerID != "" {
						inputs = nil
						break
					}
					inputs = append(inputs, port.Value)
				}
			}
			if len(inputs) > 0 && len(inputs[0]) > 0 {
				rows = append(rows, bodycodegen.AssemblyFeedback{Inputs: inputs, Expected: value.Expected})
			}
		}
	}
	return rows
}
