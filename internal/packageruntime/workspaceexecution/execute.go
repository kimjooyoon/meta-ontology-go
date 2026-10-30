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
	FillOrigins     []WorkspaceFillOrigin            `json:"fill_origins,omitempty"`
	FillOriginIssue string                           `json:"fill_origin_issue,omitempty"`
	Schema          string                           `json:"schema"`
	Program         Program                          `json:"program"`
	BodyFills       []BodyFillStep                   `json:"body_fills,omitempty"`
	SourceSHA256    string                           `json:"execution_source_sha256"`
	Composition     bodyexecution.Composition        `json:"composition"`
	Runtime         bodyexecution.CompositionRuntime `json:"runtime"`
	Scope           string                           `json:"scope"`
	Replay          *ReplayEvidence                  `json:"replay,omitempty"`
	AssemblyPolicy  *WorkspaceAssemblyPolicy         `json:"assembly_policy,omitempty"`
	PolicyHistory   []*WorkspaceAssemblyPolicy       `json:"assembly_policy_history,omitempty"`
	Continuation    *ContinuationEvidence            `json:"continuation,omitempty"`
}

type BodyFillStep struct {
	Activity          ActivityRef                 `json:"activity"`
	InputSourceSHA256 string                      `json:"input_source_sha256"`
	Generation        bodycodegen.Result          `json:"generation"`
	Plan              *bodycodegen.IRBodyFillPlan `json:"external_plan,omitempty"`
}

type ExecuteOptions struct {
	AssemblyPolicy    *packageruntime.Manifest
	AssemblyModelPath string
	GoBinary          string
	BodyFillPlans     map[string]bodycodegen.IRBodyFillPlan
	BodyFillOptions   bodycodegen.IRBodyFillOptions
	LayaEndpoint      string
	LayaAPIKey        string
	nativeExecutor    *bodyexecution.Executor
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
	policySnapshot, policy, err := prepareWorkspacePolicy(ctx, options.AssemblyPolicy)
	if err != nil {
		return Result{}, err
	}
	current := []byte(program.Source)
	current, fills, err := applyBodyFills(ctx, program, current, options)
	if err != nil {
		return Result{}, err
	}
	compositionOptions := bodyexecution.CompositionOptions{ModelPath: options.AssemblyModelPath, RecordPolicy: policy}
	if program.PureCalls != nil {
		compositionOptions.EntryActivity = program.Entry.LoweredName
	}
	composition, err := bodyexecution.GenerateCompositionWithOptions(ctx, "workspace.gooo", current, translated, compositionOptions)
	if err != nil {
		return Result{}, fmt.Errorf("generate workspace activity bodies: %w", err)
	}
	runtime, err := executeWorkspaceNative(ctx, current, composition, translated, options.GoBinary, options.nativeExecutor)
	bindWorkspaceConditionSources(program, &runtime, fills)
	bindWorkspaceLocalFileSources(program, &runtime, composition, fills)
	result := Result{Schema: "gooo/workspace-body-execution/v1", Program: program, BodyFills: fills,
		SourceSHA256: sourceSHA256(current), Composition: composition, AssemblyPolicy: policySnapshot,
		Runtime: runtime, Scope: "typed imported activity bindings; generated Go compiled and run twice; finite named expectations only"}
	result.FillOrigins, result.FillOriginIssue = collectWorkspaceFillOrigins(program, result)
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
	display := &bodyexecution.CompositionCaseDisplay{
		ActivityKeys:    make(map[string]string, len(program.Activities)),
		SubmittedInputs: make(map[string]bodyexecution.CompositionSubmittedInput),
	}
	for _, activity := range program.Activities {
		key := packageActivityKey(activity.PackagePath, activity.Activity)
		byKey[key] = activity
		display.ActivityKeys[activity.LoweredName] = key
	}
	translated := bodyexecution.CompositionCases{Schema: suite.Schema, Cases: make([]bodyexecution.CompositionCase, len(suite.Cases)), Display: display}
	for index, test := range suite.Cases {
		row := bodyexecution.CompositionCase{Inputs: make(map[string]json.RawMessage, len(test.Inputs)), Expected: make(map[string]json.RawMessage, len(test.Expected))}
		badInputKey, badInputDetail := "", ""
		for key, value := range test.Inputs {
			ref, suffix, err := resolveCaseActivity(key, byKey)
			if err != nil {
				if badInputDetail == "" || key < badInputKey {
					badInputKey, badInputDetail = key, err.Error()
				}
				continue
			}
			translatedKey := ref.LoweredName
			if suffix != "" {
				translatedKey += "." + suffix
			}
			row.Inputs[translatedKey] = append(json.RawMessage(nil), value...)
			display.SubmittedInputs[translatedKey] = bodyexecution.CompositionSubmittedInput{
				Key: key, Activity: ref.LoweredName,
			}
		}
		if badInputDetail != "" {
			return bodyexecution.CompositionCases{}, &bodyexecution.InputDiagnostic{
				Stage: "INPUT_RESOLUTION", Code: "UNKNOWN_ACTIVITY_KEY", CaseIndex: index,
				Key: badInputKey, Message: fmt.Sprintf("case %d input: %s", index, badInputDetail)}
		}
		badExpectedKey, badExpectedDetail, badExpectedCode := "", "", ""
		for key, value := range test.Expected {
			ref, suffix, err := resolveCaseActivity(key, byKey)
			if err != nil || suffix != "" {
				code := "UNKNOWN_EXPECTED_ACTIVITY"
				if err == nil {
					err = fmt.Errorf("expected outputs name activities without ports")
					code = "EXPECTED_PORT_NOT_ALLOWED"
				}
				if badExpectedDetail == "" || key < badExpectedKey {
					badExpectedKey, badExpectedDetail, badExpectedCode = key, err.Error(), code
				}
				continue
			}
			row.Expected[ref.LoweredName] = append(json.RawMessage(nil), value...)
		}
		if badExpectedDetail != "" {
			return bodyexecution.CompositionCases{}, &bodyexecution.InputDiagnostic{
				Stage: "INPUT_RESOLUTION", Code: badExpectedCode, CaseIndex: index,
				Key: badExpectedKey, Message: fmt.Sprintf("case %d expected: %s", index, badExpectedDetail)}
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
			if port := key[len(candidate)+1:]; port != "" {
				return ref, port, nil
			}
		}
	}
	return ActivityRef{}, "", fmt.Errorf("unknown package activity key %q; use <package-path>:<activity>[.<input-port>]", key)
}
