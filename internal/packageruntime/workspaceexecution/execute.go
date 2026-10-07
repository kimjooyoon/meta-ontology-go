package workspaceexecution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type Result struct {
	Schema       string                           `json:"schema"`
	Program      Program                          `json:"program"`
	BodyFills    []BodyFillStep                   `json:"body_fills,omitempty"`
	SourceSHA256 string                           `json:"execution_source_sha256"`
	Composition  bodyexecution.Composition        `json:"composition"`
	Runtime      bodyexecution.CompositionRuntime `json:"runtime"`
	Scope        string                           `json:"scope"`
	Replay       *ReplayEvidence                  `json:"replay,omitempty"`
}

type BodyFillStep struct {
	Activity          ActivityRef                 `json:"activity"`
	InputSourceSHA256 string                      `json:"input_source_sha256"`
	Generation        bodycodegen.Result          `json:"generation"`
	Plan              *bodycodegen.IRBodyFillPlan `json:"external_plan,omitempty"`
}

type ExecuteOptions struct {
	AssemblyModelPath string
	GoBinary          string
	BodyFillPlans     map[string]bodycodegen.IRBodyFillPlan
	BodyFillOptions   bodycodegen.IRBodyFillOptions
	LayaEndpoint      string
	LayaAPIKey        string
}

// ExecuteWorkspace generates every activity body in the bound workspace graph,
// compiles the emitted native Go, and observes two fresh executions against the
// same explicit finite cases.
func ExecuteWorkspace(ctx context.Context, manifest packageruntime.Manifest,
	suite bodyexecution.CompositionCases, modelPath, goBinary string) (Result, error) {
	return ExecuteWorkspaceWithOptions(ctx, manifest, suite, ExecuteOptions{AssemblyModelPath: modelPath, GoBinary: goBinary})
}

func ExecuteWorkspaceWithOptions(ctx context.Context, manifest packageruntime.Manifest,
	suite bodyexecution.CompositionCases, options ExecuteOptions) (Result, error) {
	program, err := Prepare(manifest)
	if err != nil {
		return Result{}, err
	}
	translated, err := translateCases(program, suite)
	if err != nil {
		return Result{}, err
	}
	current := []byte(program.Source)
	current, fills, err := applyBodyFills(ctx, program, current, options)
	if err != nil {
		return Result{}, err
	}
	composition, err := bodyexecution.GenerateComposition(ctx, "workspace.gooo", current, translated, options.AssemblyModelPath)
	if err != nil {
		return Result{}, fmt.Errorf("generate workspace activity bodies: %w", err)
	}
	runtime, err := bodyexecution.ExecuteComposition(ctx, "workspace.gooo", current, composition, translated, options.GoBinary)
	result := Result{Schema: "gooo/workspace-body-execution/v1", Program: program, BodyFills: fills,
		SourceSHA256: sourceSHA256(current), Composition: composition,
		Runtime: runtime, Scope: "typed imported activity bindings; generated Go compiled and run twice; finite named expectations only"}
	if err != nil {
		return result, fmt.Errorf("execute generated workspace bodies: %w", err)
	}
	result.Runtime.InputSeparation = measureWorkspaceInputs(ctx, current, result, translated)
	return result, nil
}

func sourceSHA256(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func translateCases(program Program, suite bodyexecution.CompositionCases) (bodyexecution.CompositionCases, error) {
	if suite.Schema != "gooo/body-composition-cases/v1" && suite.Schema != bodyexecution.CompositionInputsSchema {
		return bodyexecution.CompositionCases{}, fmt.Errorf("workspace execution requires gooo/body-composition-cases/v1 or gooo/body-composition-inputs/v1")
	}
	byKey := make(map[string]ActivityRef, len(program.Activities))
	for _, activity := range program.Activities {
		byKey[packageActivityKey(activity.PackagePath, activity.Activity)] = activity
	}
	translated := bodyexecution.CompositionCases{Schema: suite.Schema, Cases: make([]bodyexecution.CompositionCase, len(suite.Cases))}
	for index, test := range suite.Cases {
		row := bodyexecution.CompositionCase{Inputs: make(map[string]json.RawMessage, len(test.Inputs)), Expected: make(map[string]json.RawMessage, len(test.Expected))}
		for key, value := range test.Inputs {
			ref, suffix, err := resolveCaseActivity(key, byKey)
			if err != nil {
				return bodyexecution.CompositionCases{}, fmt.Errorf("case %d input: %w", index, err)
			}
			translatedKey := ref.LoweredName
			if suffix != "" {
				translatedKey += "." + suffix
			}
			row.Inputs[translatedKey] = append(json.RawMessage(nil), value...)
		}
		for key, value := range test.Expected {
			ref, suffix, err := resolveCaseActivity(key, byKey)
			if err != nil || suffix != "" {
				if err == nil {
					err = fmt.Errorf("expected outputs name activities without ports")
				}
				return bodyexecution.CompositionCases{}, fmt.Errorf("case %d expected: %w", index, err)
			}
			row.Expected[ref.LoweredName] = append(json.RawMessage(nil), value...)
		}
		translated.Cases[index] = row
	}
	return translated, nil
}

func resolveCaseActivity(key string, activities map[string]ActivityRef) (ActivityRef, string, error) {
	if ref, ok := activities[key]; ok {
		return ref, "", nil
	}
	for candidate, ref := range activities {
		if len(key) > len(candidate) && key[:len(candidate)] == candidate && key[len(candidate)] == '.' {
			return ref, key[len(candidate)+1:], nil
		}
	}
	return ActivityRef{}, "", fmt.Errorf("unknown package activity key %q; use <package-path>:<activity>[.<input-port>]", key)
}
