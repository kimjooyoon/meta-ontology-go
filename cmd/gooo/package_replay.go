package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageReplayUsage = "usage: gooo package replay [--json] --receipt <execution.json> (--cases <cases.json> | --inputs <inputs.json>) [--go <go-binary>] <gooo.workspace.json>"
const packageResumeUsage = "usage: gooo package resume [--json] --receipt <execution.json> --assembly-policy-workspace <policy.workspace.json> (--cases <cases.json> | --inputs <inputs.json>) [--go <go-binary>] <gooo.workspace.json>"

func runPackageReplay(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	return runPackageSaved(args, reader, stdout, stderr, false)
}

func runPackageSaved(args []string, reader SourceReader, stdout, stderr io.Writer, resume bool) int {
	args, jsonMode := parseJSONFlag(args)
	flags, manifestPath, err := parsePackageSavedArgs(args, resume)
	if err != nil {
		usage := packageReplayUsage
		if resume {
			usage = packageResumeUsage
		}
		fmt.Fprintln(stderr, usage)
		return exitUsage
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	receipt, err := executePackageSaved(ctx, reader, manifestPath, flags, resume)
	if err != nil {
		receipt.Decision, receipt.Error = "FAIL_CLOSED", err.Error()
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if writeErr := encoder.Encode(receipt); writeErr != nil {
			fmt.Fprintln(stderr, writeErr)
			return exitFailure
		}
	} else if err != nil {
		operation := "replay"
		if resume {
			operation = "resume"
		}
		fmt.Fprintf(stderr, "gooo package %s: %v\n", operation, err)
	} else if flags["--inputs"] != "" {
		return writePackageActualValues(stdout, stderr, *receipt.Result)
	} else {
		r := receipt.Result
		if resume {
			fmt.Fprintf(stdout, "continued workspace entry: %s.%s finite=%d/%d new_model_calls=%d\n",
				r.Program.Entry.PackagePath, r.Program.Entry.Activity, r.Runtime.FinitePassed, r.Runtime.FiniteTotal, r.Continuation.NewModelCalls)
		} else {
			fmt.Fprintf(stdout, "replayed workspace entry: %s.%s finite=%d/%d model_calls=%d\n",
				r.Program.Entry.PackagePath, r.Program.Entry.Activity, r.Runtime.FinitePassed, r.Runtime.FiniteTotal, r.Replay.ModelCalls)
		}
	}
	if err != nil {
		return exitFailure
	}
	return exitOK
}

func parsePackageReplayArgs(args []string) (map[string]string, string, error) {
	return parsePackageSavedArgs(args, false)
}

func parsePackageSavedArgs(args []string, resume bool) (map[string]string, string, error) {
	flags := map[string]string{"--receipt": "", "--cases": "", "--inputs": "", "--go": ""}
	if resume {
		flags["--assembly-policy-workspace"] = ""
	}
	manifest := ""
	for i := 0; i < len(args); i++ {
		if previous, ok := flags[args[i]]; ok {
			if previous != "" || i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" || strings.HasPrefix(args[i+1], "-") {
				return nil, "", fmt.Errorf("missing or repeated replay option")
			}
			flags[args[i]] = args[i+1]
			i++
		} else if manifest == "" && !strings.HasPrefix(args[i], "-") && strings.TrimSpace(args[i]) != "" {
			manifest = args[i]
		} else {
			return nil, "", fmt.Errorf("unknown replay argument")
		}
	}
	if manifest == "" || flags["--receipt"] == "" || (flags["--cases"] == "") == (flags["--inputs"] == "") {
		return nil, "", fmt.Errorf("replay needs a receipt, workspace and one input mode")
	}
	if resume && flags["--assembly-policy-workspace"] == "" {
		return nil, "", fmt.Errorf("continuation requires an explicit policy workspace")
	}
	return flags, manifest, nil
}

func executePackageReplay(ctx context.Context, reader SourceReader, manifestPath string, flags map[string]string) (packageExecutionReceipt, error) {
	return executePackageSaved(ctx, reader, manifestPath, flags, false)
}

func executePackageSaved(ctx context.Context, reader SourceReader, manifestPath string, flags map[string]string, resume bool) (packageExecutionReceipt, error) {
	receipt := packageExecutionReceipt{Schema: "gooo/workspace-body-execution-receipt/v1", Manifest: filepath.ToSlash(manifestPath)}
	manifestBytes, err := readSource(reader, manifestPath)
	if err != nil {
		return receipt, err
	}
	receipt.ManifestDigest = workspaceDigest(manifestBytes)
	manifest, err := decodeWorkspaceManifest(manifestBytes)
	if err != nil {
		return receipt, err
	}
	raw, err := readBodyExecutionFile(flags["--receipt"], 32<<20)
	if err != nil {
		return receipt, err
	}
	receipt.ReplayedFrom = workspaceDigest(raw)
	if resume {
		receipt.ContinuedFrom, receipt.ReplayedFrom = receipt.ReplayedFrom, ""
	}
	var saved packageExecutionReceipt
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		return receipt, fmt.Errorf("decode saved package execution: %w", err)
	}
	if saved.Schema != receipt.Schema || saved.Error != "" || saved.Result == nil || saved.ManifestDigest != receipt.ManifestDigest ||
		(saved.Decision != "PASS" && saved.Decision != "PROGRESS" && saved.Decision != "OBSERVED") {
		return receipt, fmt.Errorf("saved execution envelope or workspace manifest differs")
	}
	suite, digest, inputOnly, err := readPackageReplayInputs(reader, flags)
	if err != nil {
		return receipt, err
	}
	receipt.CasesDigest = digest
	if inputOnly {
		receipt.InputsDigest, receipt.CasesDigest = digest, ""
	}
	runtimeManifest, err := loadPackageSources(reader, manifestPath, manifest)
	if err != nil {
		return receipt, err
	}
	result, err := runSavedWorkspace(ctx, reader, runtimeManifest, *saved.Result, suite, flags, resume)
	receipt.Result = &result
	receipt.Decision = "PASS"
	if result.Runtime.FinitePassed < result.Runtime.FiniteTotal {
		receipt.Decision = "PROGRESS"
	}
	if inputOnly {
		receipt.Decision = "OBSERVED"
	}
	return receipt, err
}

func runSavedWorkspace(ctx context.Context, reader SourceReader, manifest packageruntime.Manifest,
	prior workspaceexecution.Result, suite bodyexecution.CompositionCases, flags map[string]string, resume bool) (workspaceexecution.Result, error) {
	if !resume {
		return workspaceexecution.ReplayWorkspace(ctx, manifest, prior, suite, flags["--go"])
	}
	policy, err := readPackageAssemblyPolicy(reader, flags["--assembly-policy-workspace"])
	if err != nil {
		return workspaceexecution.Result{}, err
	}
	return workspaceexecution.ResumeWorkspace(ctx, manifest, prior, suite, policy, flags["--go"])
}

func readPackageReplayInputs(reader SourceReader, flags map[string]string) (bodyexecution.CompositionCases, string, bool, error) {
	path, decode := flags["--cases"], bodyexecution.DecodeCompositionCases
	inputOnly := flags["--inputs"] != ""
	if inputOnly {
		path, decode = flags["--inputs"], bodyexecution.DecodeCompositionInputs
	}
	raw, err := readSource(reader, path)
	if err != nil {
		return bodyexecution.CompositionCases{}, "", inputOnly, err
	}
	suite, err := decode(raw)
	return suite, workspaceDigest(raw), inputOnly, err
}
