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

const packageExecuteUsage = "usage: gooo package execute [--json] (--cases <cases.json> | --inputs <inputs.json>) [--body-plans <plans.json>] [--assembly-model <model.json>] [--tiny-model <model.json>] [--go <go-binary>] <gooo.workspace.json>"

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
	Schema         string                     `json:"schema"`
	Decision       string                     `json:"decision"`
	Manifest       string                     `json:"manifest"`
	ManifestDigest string                     `json:"manifest_digest,omitempty"`
	CasesDigest    string                     `json:"cases_digest,omitempty"`
	InputsDigest   string                     `json:"inputs_digest,omitempty"`
	Result         *workspaceexecution.Result `json:"result,omitempty"`
	Error          string                     `json:"error,omitempty"`
}

func runPackageExecute(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	casesPath, plansPath, assemblyModelPath, tinyModelPath, goBinary, manifestPath := "", "", "", "", "", ""
	inputOnly := false
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--cases", "--inputs":
			if casesPath != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, packageExecuteUsage)
				return exitUsage
			}
			casesPath = args[index+1]
			inputOnly = args[index] == "--inputs"
			index++
		case "--assembly-model":
			if assemblyModelPath != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, packageExecuteUsage)
				return exitUsage
			}
			assemblyModelPath = args[index+1]
			index++
		case "--tiny-model":
			if tinyModelPath != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, packageExecuteUsage)
				return exitUsage
			}
			tinyModelPath = args[index+1]
			index++
		case "--body-plans":
			if plansPath != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, packageExecuteUsage)
				return exitUsage
			}
			plansPath = args[index+1]
			index++
		case "--go":
			if goBinary != "" || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, packageExecuteUsage)
				return exitUsage
			}
			goBinary = args[index+1]
			index++
		default:
			if strings.HasPrefix(args[index], "-") || manifestPath != "" {
				fmt.Fprintln(stderr, packageExecuteUsage)
				return exitUsage
			}
			manifestPath = args[index]
		}
	}
	if casesPath == "" || manifestPath == "" {
		fmt.Fprintln(stderr, packageExecuteUsage)
		return exitUsage
	}
	fail := func(cause error, manifestDigest, casesDigest string) int {
		receipt := packageExecutionReceipt{Schema: "gooo/workspace-body-execution-receipt/v1", Decision: "FAIL_CLOSED",
			Manifest: filepath.ToSlash(manifestPath), ManifestDigest: manifestDigest, CasesDigest: casesDigest, Error: cause.Error()}
		if inputOnly {
			receipt.InputsDigest, receipt.CasesDigest = receipt.CasesDigest, ""
		}
		if jsonMode {
			_ = json.NewEncoder(stdout).Encode(receipt)
		} else {
			fmt.Fprintf(stderr, "gooo package execute: %v\n", cause)
		}
		return exitFailure
	}
	manifestBytes, err := readSource(reader, manifestPath)
	if err != nil {
		return fail(err, "", "")
	}
	if int64(len(manifestBytes)) > maxInputBytes {
		return fail(inputLimitError(maxInputBytes), "", "")
	}
	manifest, err := decodeWorkspaceManifest(manifestBytes)
	if err != nil {
		return fail(err, workspaceDigest(manifestBytes), "")
	}
	caseBytes, err := readSource(reader, casesPath)
	if err != nil {
		return fail(err, workspaceDigest(manifestBytes), "")
	}
	decodeInputs := bodyexecution.DecodeCompositionCases
	if inputOnly {
		decodeInputs = bodyexecution.DecodeCompositionInputs
	}
	suite, err := decodeInputs(caseBytes)
	if err != nil {
		return fail(err, workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
	}
	plans := map[string]bodycodegen.IRBodyFillPlan{}
	if plansPath != "" {
		planBytes, readErr := readSource(reader, plansPath)
		if readErr != nil {
			return fail(readErr, workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
		}
		if len(planBytes) > 256<<10 {
			return fail(fmt.Errorf("body-fill plan set exceeds %d bytes", 256<<10), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
		}
		var planSet packageBodyFillPlanSet
		decoder := json.NewDecoder(strings.NewReader(string(planBytes)))
		decoder.DisallowUnknownFields()
		if decodeErr := decoder.Decode(&planSet); decodeErr != nil {
			return fail(fmt.Errorf("decode body-fill plan set: %w", decodeErr), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
		}
		if decodeErr := decoder.Decode(&struct{}{}); decodeErr != io.EOF {
			if decodeErr == nil {
				decodeErr = fmt.Errorf("body-fill plan set contains trailing JSON")
			}
			return fail(decodeErr, workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
		}
		if planSet.Schema != "gooo/workspace-body-fill-plans/v1" || len(planSet.Activities) == 0 || len(planSet.Activities) > 16 {
			return fail(fmt.Errorf("body-fill plan set requires schema gooo/workspace-body-fill-plans/v1 and 1..16 activity plans"), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
		}
		seenPlans := make(map[string]bool, len(planSet.Activities))
		for _, entry := range planSet.Activities {
			key := entry.PackagePath + ":" + entry.Activity
			if strings.TrimSpace(entry.PackagePath) == "" || strings.TrimSpace(entry.Activity) == "" || seenPlans[key] {
				return fail(fmt.Errorf("body-fill plan set has an empty or duplicate package activity"), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
			}
			seenPlans[key] = true
			plans[key] = entry.Plan
		}
	}
	var bodyFillOptions bodycodegen.IRBodyFillOptions
	layaEndpoint, layaAPIKey := os.Getenv("GOOO_LAYA_URL"), os.Getenv("GOOO_LAYA_API_KEY")
	if tinyModelPath != "" {
		if layaEndpoint != "" || layaAPIKey != "" {
			return fail(fmt.Errorf("--tiny-model cannot be combined with GOOO_LAYA_URL or GOOO_LAYA_API_KEY"), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
		}
		modelLoadStarted := time.Now()
		provider, loadErr := decisionroute.LoadTinyGoProvider(tinyModelPath)
		modelLoadMS := float64(time.Since(modelLoadStarted)) / float64(time.Millisecond)
		if loadErr != nil {
			return fail(fmt.Errorf("tiny_go model could not be loaded"), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
		}
		bodyFillOptions = bodycodegen.IRBodyFillOptions{TinyGoProvider: provider, TinyModelLoadMS: &modelLoadMS}
		layaEndpoint, layaAPIKey = "", ""
	}
	runtimeManifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: manifest.Entry}
	root := filepath.Dir(manifestPath)
	sourceCount, sourceBytes := 0, 0
	for _, declared := range manifest.Packages {
		pkg := packageruntime.PackageSpec{Path: declared.Path, Name: declared.Name, Imports: append([]string(nil), declared.Imports...)}
		for _, sourcePath := range declared.Sources {
			relative, pathErr := workspaceSourcePath(sourcePath)
			if pathErr != nil {
				return fail(pathErr, workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
			}
			sourceCount++
			if sourceCount > workspaceMaxSourceCount {
				return fail(fmt.Errorf("workspace declares more than %d source files", workspaceMaxSourceCount), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
			}
			filename := filepath.Join(root, relative)
			content, readErr := readSource(reader, filename)
			if readErr != nil {
				return fail(fmt.Errorf("source %q: %w", sourcePath, readErr), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
			}
			sourceBytes += len(content)
			if len(content) > workspaceMaxSourceBytes {
				return fail(fmt.Errorf("source %q exceeds %d bytes", sourcePath, workspaceMaxSourceBytes), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
			}
			if sourceBytes > workspaceMaxSourceSetSize {
				return fail(fmt.Errorf("workspace sources exceed %d bytes total", workspaceMaxSourceSetSize), workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
			}
			pkg.Sources = append(pkg.Sources, packageruntime.Source{Filename: filepath.ToSlash(relative), Content: string(content)})
		}
		runtimeManifest.Packages = append(runtimeManifest.Packages, pkg)
	}
	result, err := workspaceexecution.ExecuteWorkspaceWithOptions(context.Background(), runtimeManifest, suite, workspaceexecution.ExecuteOptions{
		AssemblyModelPath: assemblyModelPath, GoBinary: goBinary, BodyFillPlans: plans, BodyFillOptions: bodyFillOptions,
		LayaEndpoint: layaEndpoint, LayaAPIKey: layaAPIKey,
	})
	if err != nil {
		return fail(err, workspaceDigest(manifestBytes), workspaceDigest(caseBytes))
	}
	receipt := packageExecutionReceipt{Schema: "gooo/workspace-body-execution-receipt/v1", Decision: "PASS",
		Manifest: filepath.ToSlash(manifestPath), ManifestDigest: workspaceDigest(manifestBytes),
		CasesDigest: workspaceDigest(caseBytes), Result: &result}
	if inputOnly {
		receipt.Decision = "OBSERVED"
		receipt.InputsDigest, receipt.CasesDigest = receipt.CasesDigest, ""
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(receipt); err != nil {
			fmt.Fprintf(stderr, "gooo package execute: write receipt: %v\n", err)
			return exitFailure
		}
		return exitOK
	}
	if inputOnly {
		return writePackageActualValues(stdout, stderr, result)
	}
	fmt.Fprintf(stdout, "executed workspace entry: %s.%s activities=%d finite=%d/%d replayed=%t digest=%s\n",
		result.Program.Entry.PackagePath, result.Program.Entry.Activity, len(result.Program.Activities),
		result.Runtime.FinitePassed, result.Runtime.FiniteTotal, result.Runtime.RuntimeReplayed, result.Runtime.CompositionSHA256)
	return exitOK
}
