package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageExecuteUsage = "usage: gooo package execute [--json] (--cases <cases.json> | --inputs <inputs.json> | --construction-receipt <execution.json>) [--body-plans <plans.json>] [--assembly-model <model.json>] [--tiny-model <model.json>] [--go <go-binary>] <gooo.workspace.json>"

type packageBodyFillPlanSet struct {
	Schema     string                     `json:"schema"`
	Activities []packageBodyFillPlanEntry `json:"activities"`
}

type packageBodyFillPlanEntry struct {
	PackagePath string                     `json:"package_path"`
	Activity    string                     `json:"activity"`
	Plan        bodycodegen.IRBodyFillPlan `json:"plan"`
}

type packageExecutionReceipt struct {
	Schema            string                     `json:"schema"`
	Decision          string                     `json:"decision"`
	Manifest          string                     `json:"manifest"`
	ManifestDigest    string                     `json:"manifest_digest,omitempty"`
	CasesDigest       string                     `json:"cases_digest,omitempty"`
	InputsDigest      string                     `json:"inputs_digest,omitempty"`
	Result            *workspaceexecution.Result `json:"result,omitempty"`
	Error             string                     `json:"error,omitempty"`
	ReplayedFrom      string                     `json:"replayed_from_sha256,omitempty"`
	ConstructionInput *constructionInputEvidence `json:"construction_input,omitempty"`
}

func runPackageExecute(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	flags, manifestPath, err := parsePackageExecuteArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, packageExecuteUsage)
		return exitUsage
	}
	receipt, err := executePackage(context.Background(), reader, manifestPath, flags)
	if err != nil {
		receipt.Decision, receipt.Error = "FAIL_CLOSED", err.Error()
	}
	return writePackageExecution(receipt, err, jsonMode, stdout, stderr)
}

func parsePackageExecuteArgs(args []string) (map[string]string, string, error) {
	flags := map[string]string{"--cases": "", "--inputs": "", "--construction-receipt": "",
		"--body-plans": "", "--assembly-model": "", "--tiny-model": "", "--go": ""}
	manifest, inputModes := "", 0
	for i := 0; i < len(args); i++ {
		if previous, ok := flags[args[i]]; ok {
			if previous != "" || i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" || strings.HasPrefix(args[i+1], "-") {
				return nil, "", fmt.Errorf("missing or repeated execute option")
			}
			if args[i] == "--cases" || args[i] == "--inputs" || args[i] == "--construction-receipt" {
				inputModes++
			}
			flags[args[i]] = args[i+1]
			i++
		} else if manifest == "" && !strings.HasPrefix(args[i], "-") && strings.TrimSpace(args[i]) != "" {
			manifest = args[i]
		} else {
			return nil, "", fmt.Errorf("unknown execute argument")
		}
	}
	if manifest == "" || inputModes != 1 {
		return nil, "", fmt.Errorf("execute needs a workspace and one input mode")
	}
	return flags, manifest, nil
}

func executePackage(ctx context.Context, reader SourceReader, manifestPath string, flags map[string]string) (packageExecutionReceipt, error) {
	receipt := packageExecutionReceipt{Schema: "gooo/workspace-body-execution-receipt/v1", Manifest: filepath.ToSlash(manifestPath)}
	manifestBytes, err := readSource(reader, manifestPath)
	if err != nil {
		return receipt, err
	}
	if int64(len(manifestBytes)) > maxInputBytes {
		return receipt, inputLimitError(maxInputBytes)
	}
	receipt.ManifestDigest = workspaceDigest(manifestBytes)
	manifest, err := decodeWorkspaceManifest(manifestBytes)
	if err != nil {
		return receipt, err
	}
	suite, digest, observed, err := packageExecuteInputs(ctx, reader, flags, manifest.Entry)
	receipt.ConstructionInput = observed
	inputOnly := flags["--cases"] == ""
	if inputOnly {
		receipt.InputsDigest = digest
	} else {
		receipt.CasesDigest = digest
	}
	if err != nil {
		return receipt, err
	}
	options, err := packageExecuteOptions(reader, flags)
	if err != nil {
		return receipt, err
	}
	runtimeManifest, err := loadPackageSources(reader, manifestPath, manifest)
	if err != nil {
		return receipt, err
	}
	result, err := workspaceexecution.ExecuteWorkspaceWithOptions(ctx, runtimeManifest, suite, options)
	if err != nil {
		return receipt, err
	}
	receipt.Result, receipt.Decision = &result, "PASS"
	if result.Runtime.FinitePassed < result.Runtime.FiniteTotal {
		receipt.Decision = "PROGRESS"
	}
	if inputOnly {
		receipt.Decision = "OBSERVED"
	}
	return receipt, nil
}

func packageExecuteInputs(ctx context.Context, reader SourceReader, flags map[string]string, entry packageruntime.EntrySpec) (bodyexecution.CompositionCases, string, *constructionInputEvidence, error) {
	path := flags["--cases"]
	if path == "" {
		path = flags["--inputs"]
	}
	construction := flags["--construction-receipt"] != ""
	if construction {
		path = flags["--construction-receipt"]
	}
	raw, err := readPackageInputBytes(reader, path, construction)
	if err != nil {
		return bodyexecution.CompositionCases{}, "", nil, err
	}
	var observed *constructionInputEvidence
	if construction {
		raw, observed, err = packageConstructionInputs(ctx, raw, entry)
		if err != nil {
			return bodyexecution.CompositionCases{}, "", nil, err
		}
	}
	decode := bodyexecution.DecodeCompositionCases
	if flags["--cases"] == "" {
		decode = bodyexecution.DecodeCompositionInputs
	}
	suite, err := decode(raw)
	return suite, workspaceDigest(raw), observed, err
}

func packageExecuteOptions(reader SourceReader, flags map[string]string) (workspaceexecution.ExecuteOptions, error) {
	options := workspaceexecution.ExecuteOptions{AssemblyModelPath: flags["--assembly-model"], GoBinary: flags["--go"],
		LayaEndpoint: os.Getenv("GOOO_LAYA_URL"), LayaAPIKey: os.Getenv("GOOO_LAYA_API_KEY")}
	plans, err := readPackageBodyPlans(reader, flags["--body-plans"])
	if err != nil {
		return options, err
	}
	options.BodyFillPlans = plans
	if path := flags["--tiny-model"]; path != "" {
		if options.LayaEndpoint != "" || options.LayaAPIKey != "" {
			return options, fmt.Errorf("--tiny-model cannot be combined with GOOO_LAYA_URL or GOOO_LAYA_API_KEY")
		}
		started := time.Now()
		provider, loadErr := decisionroute.LoadTinyGoProvider(path)
		loadMS := float64(time.Since(started)) / float64(time.Millisecond)
		if loadErr != nil {
			return options, fmt.Errorf("tiny_go model could not be loaded")
		}
		options.BodyFillOptions = bodycodegen.IRBodyFillOptions{TinyGoProvider: provider, TinyModelLoadMS: &loadMS}
	}
	return options, nil
}

func readPackageBodyPlans(reader SourceReader, path string) (map[string]bodycodegen.IRBodyFillPlan, error) {
	plans := map[string]bodycodegen.IRBodyFillPlan{}
	if path == "" {
		return plans, nil
	}
	raw, err := readSource(reader, path)
	if err != nil {
		return nil, err
	}
	if len(raw) > 256<<10 {
		return nil, fmt.Errorf("body-fill plan set exceeds %d bytes", 256<<10)
	}
	var planSet packageBodyFillPlanSet
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&planSet); err != nil {
		return nil, fmt.Errorf("decode body-fill plan set: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("body-fill plan set contains trailing JSON")
		}
		return nil, err
	}
	if planSet.Schema != "gooo/workspace-body-fill-plans/v1" || len(planSet.Activities) == 0 || len(planSet.Activities) > 16 {
		return nil, fmt.Errorf("body-fill plan set requires schema gooo/workspace-body-fill-plans/v1 and 1..16 activity plans")
	}
	for _, entry := range planSet.Activities {
		key := entry.PackagePath + ":" + entry.Activity
		if _, exists := plans[key]; strings.TrimSpace(entry.PackagePath) == "" || strings.TrimSpace(entry.Activity) == "" || exists {
			return nil, fmt.Errorf("body-fill plan set has an empty or duplicate package activity")
		}
		plans[key] = entry.Plan
	}
	return plans, nil
}

func writePackageExecution(receipt packageExecutionReceipt, cause error, jsonMode bool, stdout, stderr io.Writer) int {
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(receipt); err != nil {
			fmt.Fprintf(stderr, "gooo package execute: write receipt: %v\n", err)
			return exitFailure
		}
	} else if cause != nil {
		fmt.Fprintf(stderr, "gooo package execute: %v\n", cause)
	} else if receipt.Decision == "OBSERVED" {
		return writePackageActualValues(stdout, stderr, *receipt.Result)
	} else {
		result := receipt.Result
		fmt.Fprintf(stdout, "executed workspace entry: %s.%s activities=%d finite=%d/%d replayed=%t digest=%s\n",
			result.Program.Entry.PackagePath, result.Program.Entry.Activity, len(result.Program.Activities),
			result.Runtime.FinitePassed, result.Runtime.FiniteTotal, result.Runtime.RuntimeReplayed, result.Runtime.CompositionSHA256)
	}
	if cause != nil {
		return exitFailure
	}
	return exitOK
}
