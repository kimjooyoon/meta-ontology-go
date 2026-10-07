package workspaceexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// ObserveConstruction reconstructs source-owned fills followed by composition
// before returning finite construction observations. Native history is not replayed.
func ObserveConstruction(ctx context.Context, prior Result) ([]bodyexecution.ConstructionObservation, error) {
	if ctx == nil || prior.Schema != "gooo/workspace-body-execution/v1" {
		return nil, fmt.Errorf("construction observation requires a context and workspace execution")
	}
	current := []byte(prior.Program.Source)
	var rows []bodyexecution.ConstructionObservation
	seen := make(map[string]bool)
	for _, step := range prior.BodyFills {
		activity := step.Activity.LoweredName
		if seen[activity] || step.InputSourceSHA256 != sourceSHA256(current) ||
			step.Generation.Report.Activity != activity || step.Generation.Report.BodyFill == nil {
			return nil, fmt.Errorf("construction body fill has a changed source, activity or profile")
		}
		seen[activity] = true
		replayed, err := bodycodegen.RealizeSourceAssembly(ctx, "workspace.gooo", current, step.Generation)
		if err != nil {
			return nil, fmt.Errorf("construction observation requires replayable source-owned fills: %w", err)
		}
		current = []byte(replayed.Source)
		rows = append(rows, bodyexecution.ObserveVerifiedFill(step.Generation.Report)...)
	}
	if sourceSHA256(current) != prior.SourceSHA256 {
		return nil, fmt.Errorf("construction input source differs from the saved execution source")
	}
	composed, err := bodyexecution.ObserveConstruction(ctx, current, prior.Composition)
	if err != nil {
		return nil, err
	}
	rows = append(rows, composed...)
	if len(rows) == 0 {
		return nil, fmt.Errorf("saved workspace has no construction observations")
	}
	return rows, nil
}
