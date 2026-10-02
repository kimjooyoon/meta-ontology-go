package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

const bodyCodegenUsage = "usage: gooo body-codegen [--json] " +
	"[--sample-seed <seed> | --fill-plan <plan.json> [--tiny-model <model.json>] | --fill-search <plan.json> | " +
	"--path-plan <plan.json> [--path-model <model.json>] [--path-step-attempts <1..64>] " +
	"[--path-diagnosis <diagnosis.json>] " +
	"[--path-feedback-rounds <1..16> [--path-feedback-ci <hint.json>] [--path-feedback-unfixed]]] " +
	"--activity <name> <file.gooo>"
const tinyModelDiagnosticLabel = "<tiny_model>"

func runBodyCodegen(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runBodyCodegenContext(ctx, args, reader, stdout, stderr)
}

func runBodyCodegenContext(ctx context.Context, args []string, reader SourceReader, stdout, stderr io.Writer) int {
	jsonMode := false
	activity := ""
	sampleSeed := ""
	sampleSeedSet := false
	fillPlanPath := ""
	searchPlanPath := ""
	tinyModelPath := ""
	pathPlanPath := ""
	pathModelPath := ""
	pathStepAttempts := 0
	pathFeedbackRounds := 0
	pathFeedbackCI := ""
	pathFeedbackUnfixed := false
	pathDiagnosisPath := ""
	filename := ""
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--json":
			jsonMode = true
		case "--activity":
			if index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			activity = args[index+1]
			index++
		case "--sample-seed":
			if sampleSeedSet || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" ||
				strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			sampleSeed = args[index+1]
			sampleSeedSet = true
			index++
		case "--fill-plan":
			if fillPlanPath != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" ||
				strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			fillPlanPath = args[index+1]
			index++
		case "--fill-search":
			if searchPlanPath != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" ||
				strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			searchPlanPath = args[index+1]
			index++
		case "--tiny-model":
			if tinyModelPath != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" ||
				strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			tinyModelPath = args[index+1]
			index++
		case "--path-step-attempts":
			if pathStepAttempts != 0 || index+1 >= len(args) {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			value, err := strconv.Atoi(args[index+1])
			if err != nil || value < 1 || value > 64 {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			pathStepAttempts = value
			index++
		case "--path-feedback-rounds":
			if pathFeedbackRounds != 0 || index+1 >= len(args) {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			value, err := strconv.Atoi(args[index+1])
			if err != nil || value < 1 || value > 16 {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			pathFeedbackRounds = value
			index++
		case "--path-feedback-ci":
			if pathFeedbackCI != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			pathFeedbackCI = args[index+1]
			index++
		case "--path-feedback-unfixed":
			if pathFeedbackUnfixed {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			pathFeedbackUnfixed = true
		case "--path-plan", "--path-model", "--path-diagnosis":
			target := &pathPlanPath
			if args[index] == "--path-model" {
				target = &pathModelPath
			}
			if args[index] == "--path-diagnosis" {
				target = &pathDiagnosisPath
			}
			if *target != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" ||
				strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			*target = args[index+1]
			index++
		default:
			if strings.HasPrefix(args[index], "-") || filename != "" {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			filename = args[index]
		}
	}
	if activity == "" || filename == "" || (sampleSeedSet && (fillPlanPath != "" || searchPlanPath != "")) ||
		(fillPlanPath != "" && searchPlanPath != "") ||
		(tinyModelPath != "" && (fillPlanPath == "" || searchPlanPath != "" || sampleSeedSet)) {
		fmt.Fprintln(stderr, bodyCodegenUsage)
		return exitUsage
	}
	if (pathPlanPath != "" && (sampleSeedSet || fillPlanPath != "" || searchPlanPath != "" || tinyModelPath != "")) ||
		((pathModelPath != "" || pathStepAttempts != 0 || pathFeedbackRounds != 0 || pathFeedbackCI != "" ||
			pathDiagnosisPath != "") && pathPlanPath == "") ||
		(pathFeedbackRounds != 0 && (pathModelPath == "" || pathStepAttempts == 0)) ||
		((pathFeedbackCI != "" || pathFeedbackUnfixed) && pathFeedbackRounds == 0) {
		fmt.Fprintln(stderr, bodyCodegenUsage)
		return exitUsage
	}
	if tinyModelPath != "" && (strings.TrimSpace(os.Getenv("GOOO_LAYA_URL")) != "" ||
		strings.TrimSpace(os.Getenv("GOOO_LAYA_API_KEY")) != "") {
		fmt.Fprintln(stderr, "gooo: --tiny-model cannot be combined with configured GOOO_LAYA_URL or GOOO_LAYA_API_KEY")
		return exitUsage
	}
	source, err := reader.ReadFile(filename)
	if err != nil {
		return reportBodyCodegenFailure(jsonMode, filename, activity, nil, err, stdout, stderr)
	}
	var result bodycodegen.Result
	if pathPlanPath != "" {
		planBytes, readErr := reader.ReadFile(pathPlanPath)
		if readErr != nil {
			return reportBodyCodegenFailure(jsonMode, pathPlanPath, activity, planBytes, readErr, stdout, stderr)
		}
		document, decodeErr := pathplan.DecodeDocument(planBytes)
		if decodeErr != nil {
			return reportBodyCodegenFailure(jsonMode, pathPlanPath, activity, planBytes, decodeErr, stdout, stderr)
		}
		if pathDiagnosisPath != "" {
			options := bodycodegen.TypedPathOptions{StepAttempts: pathStepAttempts, FeedbackRounds: pathFeedbackRounds,
				FeedbackUnfixed: pathFeedbackUnfixed}
			result, err = generateWithPathDiagnosis(ctx, reader, filename, source, activity, document, pathModelPath,
				pathDiagnosisPath, pathFeedbackCI, options)
		} else if pathFeedbackRounds != 0 {
			var ci *pathplan.CIHint
			if pathFeedbackCI != "" {
				raw, readErr := reader.ReadFile(pathFeedbackCI)
				if readErr != nil {
					return reportBodyCodegenFailure(jsonMode, pathFeedbackCI, activity, raw, readErr, stdout, stderr)
				}
				ci, decodeErr = decodePathCIHint(raw)
				if decodeErr != nil {
					return reportBodyCodegenFailure(jsonMode, pathFeedbackCI, activity, raw, decodeErr, stdout, stderr)
				}
			}
			generate := bodycodegen.GenerateWithTypedPathFeedback
			if pathFeedbackUnfixed {
				generate = bodycodegen.GenerateWithTypedPathUnfixedFeedback
			}
			result, err = generate(ctx, filename, source, activity, document, pathModelPath,
				pathStepAttempts, pathFeedbackRounds, ci)
		} else if pathStepAttempts == 0 {
			result, err = bodycodegen.GenerateWithTypedPaths(ctx, filename, source, activity, document, pathModelPath)
		} else {
			result, err = bodycodegen.GenerateWithTypedPathBatches(ctx, filename, source, activity,
				document, pathModelPath, pathStepAttempts)
		}
	} else if fillPlanPath != "" || searchPlanPath != "" {
		planPath := fillPlanPath
		if searchPlanPath != "" {
			planPath = searchPlanPath
		}
		planBytes, readErr := reader.ReadFile(planPath)
		if readErr != nil {
			return reportBodyCodegenFailure(jsonMode, planPath, activity, planBytes, readErr, stdout, stderr)
		}
		var plan bodycodegen.IRBodyFillPlan
		var searchPlan bodycodegen.IRBodySearchPlan
		var target any = &plan
		if searchPlanPath != "" {
			target = &searchPlan
		}
		decoder := json.NewDecoder(strings.NewReader(string(planBytes)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			return reportBodyCodegenFailure(jsonMode, planPath, activity, planBytes, err, stdout, stderr)
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			if err == nil {
				err = fmt.Errorf("multiple JSON values in IR body-fill plan")
			}
			return reportBodyCodegenFailure(jsonMode, planPath, activity, planBytes, err, stdout, stderr)
		}
		if searchPlanPath != "" {
			result, err = bodycodegen.GenerateWithIRBodySearch(ctx, filename, source, activity,
				searchPlan, os.Getenv("GOOO_LAYA_URL"), os.Getenv("GOOO_LAYA_API_KEY"))
		} else {
			if tinyModelPath != "" {
				if plan.ProviderModel != "" {
					return reportBodyCodegenFailure(jsonMode, planPath, activity, planBytes,
						fmt.Errorf("--tiny-model cannot be combined with the plan provider_model selector"), stdout, stderr)
				}
				modelLoadStarted := time.Now()
				provider, loadErr := decisionroute.LoadTinyGoProvider(tinyModelPath)
				modelLoadMS := float64(time.Since(modelLoadStarted)) / float64(time.Millisecond)
				if loadErr != nil {
					return reportBodyCodegenFailure(jsonMode, tinyModelDiagnosticLabel, activity, planBytes, loadErr, stdout, stderr)
				}
				result, err = bodycodegen.GenerateWithIRBodyFillWithOptions(ctx, filename, source, activity,
					plan, "", "", bodycodegen.IRBodyFillOptions{TinyGoProvider: provider, TinyModelLoadMS: &modelLoadMS})
			} else {
				result, err = bodycodegen.GenerateWithIRBodyFillWithOptions(ctx, filename, source, activity,
					plan, os.Getenv("GOOO_LAYA_URL"), os.Getenv("GOOO_LAYA_API_KEY"), bodycodegen.IRBodyFillOptions{})
			}
		}
	} else {
		result, err = bodycodegen.GenerateWithPlannerAndSampleSeed(
			ctx, filename, source, activity,
			os.Getenv("GOOO_LAYA_URL"), os.Getenv("GOOO_LAYA_API_KEY"), sampleSeed,
		)
	}
	if err != nil {
		return reportBodyCodegenFailure(jsonMode, filename, activity, source, err, stdout, stderr)
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(result); err != nil {
			fmt.Fprintf(stderr, "gooo: body-codegen output: %v\n", err)
			return exitFailure
		}
		return exitOK
	}
	if _, err := io.WriteString(stdout, result.Source); err != nil {
		fmt.Fprintf(stderr, "gooo: body-codegen output: %v\n", err)
		return exitFailure
	}
	return exitOK
}

func reportBodyCodegenFailure(jsonMode bool, filename, activity string, source []byte, cause error, stdout, stderr io.Writer) int {
	if jsonMode {
		completeness := bodycodegen.FailureCompletenessReceipt(activity, source, cause.Error())
		var searchReceipt *bodycodegen.IRBodySearchReceipt
		var pathReceipt *bodycodegen.BodyPathReceipt
		if pathError, ok := errors.AsType[*bodycodegen.BodyPathError](cause); ok {
			pathReceipt = pathError.Receipt
			completeness = bodycodegen.PathFailureCompletenessReceipt(activity, source, cause.Error(), pathReceipt)
		}
		if searchError, ok := errors.AsType[*bodycodegen.IRBodySearchError](cause); ok {
			searchReceipt = searchError.Receipt
			completeness = bodycodegen.SearchFailureCompletenessReceipt(activity, source, cause.Error(), searchReceipt)
		}
		payload := struct {
			Schema              string                           `json:"schema"`
			Decision            string                           `json:"decision"`
			Source              string                           `json:"source"`
			PlanSHA256          string                           `json:"plan_sha256"`
			CompilerSourceSHA   string                           `json:"compiler_source_sha"`
			RepositoryWrites    int                              `json:"repository_writes"`
			Error               string                           `json:"error"`
			CompletenessReceipt *bodycodegen.CompletenessReceipt `json:"completeness_receipt"`
			BodySearch          *bodycodegen.IRBodySearchReceipt `json:"body_search,omitempty"`
			BodyPaths           *bodycodegen.BodyPathReceipt     `json:"body_paths,omitempty"`
		}{
			Schema: "gooo/body-codegen-report/v3", Decision: "FAIL_CLOSED",
			PlanSHA256:        stringScopeValue(completeness.Scope, "plan_sha256"),
			CompilerSourceSHA: stringScopeValue(completeness.Scope, "compiler_source_sha"),
			RepositoryWrites:  0, Error: cause.Error(), CompletenessReceipt: completeness,
			BodySearch: searchReceipt, BodyPaths: pathReceipt,
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(payload); err != nil {
			fmt.Fprintf(stderr, "gooo: body-codegen output: %v\n", err)
		}
		return exitFailure
	}
	fmt.Fprintf(stderr, "gooo: %s: body-codegen: %v\n", filename, cause)
	return exitFailure
}

func stringScopeValue(scope map[string]any, key string) string {
	value, _ := scope[key].(string)
	return value
}
