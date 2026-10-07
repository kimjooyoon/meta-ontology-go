package workspaceexecution

import (
	"context"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// ObserveConstruction reconstructs source or externally planned fills, then composition
// before returning finite construction observations. Native history is not replayed.
func ObserveConstruction(ctx context.Context, prior Result) ([]bodyexecution.ConstructionObservation, error) {
	if ctx == nil || prior.Schema != "gooo/workspace-body-execution/v1" {
		return nil, fmt.Errorf("construction observation requires a context and workspace execution")
	}
	current := []byte(prior.Program.Source)
	var rows []bodyexecution.ConstructionObservation
	lastIndex := -1
	for _, step := range prior.BodyFills {
		activity := step.Activity.LoweredName
		index := slices.Index(prior.Program.Activities, step.Activity)
		if index <= lastIndex || step.InputSourceSHA256 != sourceSHA256(current) ||
			step.Generation.Report.Activity != activity || step.Generation.Report.BodyFill == nil {
			return nil, fmt.Errorf("construction body fill has a changed source, activity or profile")
		}
		lastIndex = index
		replayed, observed, err := observeWorkspaceFill(ctx, current, step)
		if err != nil {
			return nil, err
		}
		current = []byte(replayed)
		rows = append(rows, observed...)
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

func observeWorkspaceFill(ctx context.Context, source []byte, step BodyFillStep) (string, []bodyexecution.ConstructionObservation, error) {
	spec, err := bodycodegen.SourceAssembly(ctx, "workspace.gooo", source, step.Activity.LoweredName)
	if err != nil {
		return "", nil, err
	}
	if step.Plan != nil {
		if spec != nil {
			return "", nil, fmt.Errorf("construction body fill has both a source contract and an external plan")
		}
		completed, err := bodycodegen.ReplayIRBodyFill(ctx, "workspace.gooo", source, *step.Plan, step.Generation)
		if err != nil {
			return "", nil, err
		}
		return completed, bodyexecution.ObserveVerifiedExternalFill(step.Generation.Report), nil
	}
	if !bodycodegen.IsSourceIRBodyFill(spec) {
		return "", nil, fmt.Errorf("external construction observation requires its saved plan; regenerate older receipts with the original plan")
	}
	completed, err := bodycodegen.RealizeSourceAssembly(ctx, "workspace.gooo", source, step.Generation)
	if err != nil {
		return "", nil, err
	}
	return completed.Source, bodyexecution.ObserveVerifiedFill(step.Generation.Report), nil
}
