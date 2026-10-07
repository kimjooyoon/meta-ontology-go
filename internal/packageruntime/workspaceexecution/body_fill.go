package workspaceexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func applyBodyFills(ctx context.Context, program Program, current []byte, options ExecuteOptions) ([]byte, []BodyFillStep, error) {
	fills := make([]BodyFillStep, 0, len(options.BodyFillPlans)+len(program.sourceFillSpecs))
	knownPlans := make(map[string]bool, len(options.BodyFillPlans))
	knownSourcePlans := make(map[string]bool, len(program.sourceFillSpecs))
	bodyFillOptions := options.BodyFillOptions
	for _, activity := range program.Activities {
		key := packageActivityKey(activity.PackagePath, activity.Activity)
		plan, exists := options.BodyFillPlans[key]
		sourceSpec := program.sourceFillSpecs[key]
		if exists && sourceSpec != nil {
			return nil, nil, fmt.Errorf("activity %s declares both a Gooo source fill plan and an external body-fill plan", key)
		}
		if !exists && sourceSpec == nil {
			continue
		}
		if exists {
			var err error
			plan, err = program.recordNames.rewriteFillPlan(activity.PackagePath, plan)
			if err != nil {
				return nil, nil, err
			}
		}
		generation, err := generateActivityBodyFill(ctx, current, activity, plan, sourceSpec, options, bodyFillOptions)
		if err != nil {
			return nil, nil, fmt.Errorf("activity %s body fill: %w", key, err)
		}
		if sourceSpec != nil {
			knownSourcePlans[key] = true
		} else {
			knownPlans[key] = true
		}
		bodyFillOptions.TinyModelLoadMS = nil
		if generation.GoooSource == "" || generation.Report.BodyFill == nil {
			return nil, nil, fmt.Errorf("activity %s body fill returned no replayable Gooo source", key)
		}
		fills = append(fills, BodyFillStep{Activity: activity, InputSourceSHA256: sourceSHA256(current), Generation: generation})
		current = []byte(generation.GoooSource)
	}
	if len(knownPlans) != len(options.BodyFillPlans) {
		return nil, nil, fmt.Errorf("body-fill plans contain an activity not declared in the workspace")
	}
	if len(knownSourcePlans) != len(program.sourceFillSpecs) {
		return nil, nil, fmt.Errorf("a Gooo source body-fill plan targets an activity outside the executable workspace path")
	}
	if options.BodyFillOptions.TinyGoProvider != nil && len(fills) == 0 {
		return nil, nil, fmt.Errorf("tiny_go model was supplied but the workspace declares no body-fill plan")
	}
	return current, fills, nil
}

func generateActivityBodyFill(ctx context.Context, current []byte, activity ActivityRef,
	plan bodycodegen.IRBodyFillPlan, sourceSpec *assemblyspec.Spec, options ExecuteOptions,
	bodyFillOptions bodycodegen.IRBodyFillOptions) (bodycodegen.Result, error) {
	if sourceSpec != nil {
		return bodycodegen.GenerateWithSourceIRBodyFill(ctx, "workspace.gooo", current,
			activity.LoweredName, sourceSpec, options.LayaEndpoint, options.LayaAPIKey, bodyFillOptions)
	}
	return bodycodegen.GenerateWithIRBodyFillWithOptions(ctx, "workspace.gooo", current,
		activity.LoweredName, plan, options.LayaEndpoint, options.LayaAPIKey, bodyFillOptions)
}
