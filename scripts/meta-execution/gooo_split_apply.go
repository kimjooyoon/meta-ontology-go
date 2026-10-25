package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func applyAndVerifyGoooSplit(root string, environment []string, proposalPath string, request workspacesplit.Request,
	proposal goooSplitReceipt, plan generation.Plan, action generation.Action, trace metaExecutionTrace,
	pass string, materialized operationMaterialization) (operationMaterialization, *operationError) {
	args := []string{"package", "split-source", "--apply", proposalPath, "--json", filepath.Join(root, filepath.FromSlash(request.Workspace))}
	applied, err := runGoooSplitCommand(root, environment, args, &trace, pass, "apply")
	materialized.Evaluator = applied.Observation
	if err != nil || applied.Observation.ExitCode != 0 {
		return materialized, goooSplitProcessFailure("apply", applied, err)
	}
	receipt, err := readGoooSplitReceipt(applied.Stdout)
	if err != nil || receipt.Decision != "APPLIED" || receipt.Application == nil || receipt.Application.State != "OUTPUT_VERIFIED" ||
		!reflect.DeepEqual(proposal.Result, receipt.Result) || !reflect.DeepEqual(proposal.Selection, receipt.Selection) ||
		!reflect.DeepEqual(proposal.Registration, receipt.Registration) || !reflect.DeepEqual(proposal.Discovery, receipt.Discovery) {
		return materialized, goooSplitFailure("read-application", "WORKSPACE_SPLIT_APPLICATION_MISMATCH", err)
	}
	if err := verifyInstalledGoooSplit(root, request, receipt); err != nil {
		return materialized, goooSplitFailure("read-installed-files", "WORKSPACE_SPLIT_OUTPUT_MISMATCH", err)
	}
	runtime, runtimeProcess, failure := verifyGoooSplitRuntime(root, environment, request, trace, pass)
	materialized.Verifier = runtimeProcess.Observation
	if failure != nil {
		return materialized, failure
	}
	native, err := verifyGoooSplitNative(root, environment, trace, pass)
	materialized.Evaluator = native.Observation
	if err != nil {
		return materialized, goooSplitFailure("native-conformance", "WORKSPACE_SPLIT_NATIVE_CONFORMANCE_FAILED", err)
	}
	canonical, err := json.Marshal(goooSplitReplay{Request: request, Receipt: json.RawMessage(canonicalProcessBytes(root, applied.Stdout)),
		Runtime: runtime, Native: string(native.Stdout), Executor: splitReplayProcessFrom(materialized.Executor),
		Evaluator: splitReplayProcessFrom(materialized.Evaluator), Verifier: splitReplayProcessFrom(materialized.Verifier)})
	if err != nil {
		return materialized, goooSplitFailure("encode-instance", "WORKSPACE_SPLIT_EVIDENCE_FAILED", err)
	}
	return finishGoooSplitMaterialization(plan, action, receipt, canonical, materialized)
}

func verifyInstalledGoooSplit(root string, request workspacesplit.Request, receipt goooSplitReceipt) error {
	for _, file := range receipt.Result.Files {
		if !fs.ValidPath(file.Filename) {
			return fmt.Errorf("noncanonical source output")
		}
		name := filepath.Join(root, filepath.Dir(filepath.FromSlash(request.Workspace)), filepath.FromSlash(file.Filename))
		raw, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if string(raw) != file.Source || digestBytes(raw) != file.SHA256 {
			return fmt.Errorf("installed source differs")
		}
	}
	for _, output := range receipt.Registration.Outputs {
		if !fs.ValidPath(output.Path) {
			return fmt.Errorf("noncanonical registry output")
		}
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(output.Path)))
		if err != nil {
			return err
		}
		if string(raw) != output.Source || digestBytes(raw) != output.SHA256 {
			return fmt.Errorf("installed registry differs")
		}
	}
	return nil
}

func verifyGoooSplitNative(root string, environment []string, trace metaExecutionTrace, pass string) (processResult, error) {
	command := []string{"go", "test", "-count=1", "./internal/meta/languagereadiness/languagesyntax/conformance",
		"./internal/meta/languageassurance/verticalsliceclosureshadow"}
	process, err := runProcessObserved(root, environment, command, command, &trace, pass, "native-verifier")
	if err != nil || process.Observation.ExitCode != 0 {
		return process, fmt.Errorf("native registry conformance failed: %w", err)
	}
	normalized, err := normalizeRegistrationVerifier(process.Stdout)
	if err != nil {
		return process, err
	}
	process.Stdout = normalized
	process.Observation.StdoutBytes, process.Observation.StdoutDigest = len(normalized), digestBytes(normalized)
	return process, nil
}
