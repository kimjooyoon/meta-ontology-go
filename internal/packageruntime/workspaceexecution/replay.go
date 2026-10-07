package workspaceexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type ReplayEvidence struct {
	Schema            string `json:"schema"`
	PriorResultSHA256 string `json:"prior_result_sha256"`
	BodyFillsReplayed int    `json:"body_fills_replayed"`
	ModelCalls        int    `json:"model_calls"`
	Scope             string `json:"scope"`
}

// ReplayWorkspace reconstructs the current workspace and previously selected
// bodies, then executes those bodies against new inputs. Prior runtime results
// remain historical evidence; this invocation performs two fresh native runs.
func ReplayWorkspace(ctx context.Context, manifest packageruntime.Manifest, prior Result,
	suite bodyexecution.CompositionCases, goBinary string) (Result, error) {
	if ctx == nil {
		return Result{}, fmt.Errorf("workspace replay requires a context")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if prior.Schema != "gooo/workspace-body-execution/v1" {
		return Result{}, fmt.Errorf("workspace replay requires a saved execution result")
	}
	program, err := Prepare(manifest)
	if err != nil {
		return Result{}, err
	}
	currentProgram, _ := json.Marshal(program)
	savedProgram, _ := json.Marshal(prior.Program)
	if !bytes.Equal(currentProgram, savedProgram) {
		return Result{}, fmt.Errorf("saved workspace differs from current sources, dependencies, entry or lowered graph")
	}
	translated, err := translateCases(program, suite)
	if err != nil {
		return Result{}, err
	}
	current, err := replayWorkspaceFills(ctx, program, prior.BodyFills)
	if err != nil {
		return Result{}, err
	}
	if sourceSHA256(current) != prior.SourceSHA256 {
		return Result{}, fmt.Errorf("saved workspace execution source differs")
	}
	if err := verifyWorkspacePolicy(ctx, prior); err != nil {
		return Result{}, err
	}
	encoded, err := json.Marshal(prior)
	if err != nil {
		return Result{}, err
	}
	result := Result{Schema: prior.Schema, Program: program, BodyFills: prior.BodyFills,
		SourceSHA256: prior.SourceSHA256, Composition: prior.Composition, AssemblyPolicy: prior.AssemblyPolicy,
		Replay: &ReplayEvidence{Schema: "gooo/workspace-body-replay/v1", PriorResultSHA256: sourceSHA256(encoded),
			BodyFillsReplayed: len(prior.BodyFills), Scope: "saved decisions and source-bound candidates reconstructed without inference; old runtime observations are not reused"},
		Scope: "saved workspace construction with two fresh native executions; current finite expectations or input-only observations"}
	result.Runtime, err = bodyexecution.ExecuteComposition(ctx, "workspace.gooo", current, prior.Composition, translated, goBinary)
	if err == nil {
		result.Runtime.InputSeparation = measureWorkspaceInputs(ctx, current, result, translated)
	}
	return result, err
}

func replayWorkspaceFills(ctx context.Context, program Program, fills []BodyFillStep) ([]byte, error) {
	if len(fills) > len(program.Activities) {
		return nil, fmt.Errorf("saved workspace has too many body-fill steps")
	}
	current := []byte(program.Source)
	lastIndex, sourceFills := -1, 0
	for _, step := range fills {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		index := -1
		for i, activity := range program.Activities {
			if activity == step.Activity {
				index = i
			}
		}
		if index <= lastIndex || step.InputSourceSHA256 != sourceSHA256(current) || step.Generation.Report.Activity != step.Activity.LoweredName {
			return nil, fmt.Errorf("saved workspace body-fill order or activity binding differs")
		}
		lastIndex = index
		key := packageActivityKey(step.Activity.PackagePath, step.Activity.Activity)
		var completed string
		var err error
		if program.sourceFillSpecs[key] != nil {
			if step.Plan != nil {
				return nil, fmt.Errorf("source fill cannot also carry an external plan")
			}
			var realized bodycodegen.Realization
			realized, err = bodycodegen.RealizeSourceAssembly(ctx, "workspace.gooo", current, step.Generation)
			completed = realized.Source
			sourceFills++
		} else {
			if step.Plan == nil {
				return nil, fmt.Errorf("external body-fill replay requires its saved plan; regenerate older receipts once with the original plan")
			}
			completed, err = bodycodegen.ReplayIRBodyFill(ctx, "workspace.gooo", current, *step.Plan, step.Generation)
		}
		if err != nil {
			return nil, fmt.Errorf("replay workspace activity %s: %w", key, err)
		}
		current = []byte(completed)
	}
	if sourceFills != len(program.sourceFillSpecs) {
		return nil, fmt.Errorf("saved workspace omits a source-declared body fill")
	}
	return current, nil
}
