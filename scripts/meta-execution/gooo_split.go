package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func executeGoooSplit(workspace, gitDir string, plan generation.Plan, action generation.Action, trace metaExecutionTrace) (operationMaterialization, *operationError) {
	if action.WorkspaceSplitRequest == nil {
		return operationMaterialization{}, newOperationError("OPERATION_INPUT", "bind-workspace-scope",
			"WORKSPACE_SPLIT_REQUEST_MISSING", "DIRECT_MISSING", "bind-exact-workspace-consumer-scope")
	}
	if !generation.ValidRegistrationActionInput(action) {
		return operationMaterialization{}, goooSplitFailure("bind-request", "WORKSPACE_SPLIT_REQUEST_MISMATCH", nil)
	}
	snapshot, err := readOnlyGitSnapshot(gitDir, plan.HeadSHA)
	if err != nil {
		return operationMaterialization{}, goooSplitFailure("snapshot", "EXACT_GIT_INPUT_UNAVAILABLE", err)
	}
	defer cleanupGitSnapshot(filepath.Dir(snapshot))
	replay, err := newMetaReplayWorkspace()
	if err != nil {
		return operationMaterialization{}, goooSplitFailure("workspace", "WORKSPACE_UNAVAILABLE", err)
	}
	defer replay.close()
	first, failure := materializeGoooSplit(snapshot, replay.path, plan, action, trace, "first")
	if failure != nil {
		return first, failure
	}
	second, failure := materializeGoooSplit(snapshot, replay.path, plan, action, trace, "replay")
	if failure != nil {
		return second, failure
	}
	if !bytes.Equal(first.Canonical, second.Canonical) || first.InstanceDigest != second.InstanceDigest {
		return first, goooSplitFailure("compare-replay", "WORKSPACE_SPLIT_REPLAY_MISMATCH", nil)
	}
	first.Verifier = second.Verifier
	return first, nil
}

func materializeGoooSplit(gitDir, root string, plan generation.Plan, action generation.Action,
	trace metaExecutionTrace, pass string) (operationMaterialization, *operationError) {
	var materialized operationMaterialization
	if err := restoreGoooSplitSnapshot(root, gitDir, plan.HeadSHA); err != nil {
		return materialized, goooSplitFailure("restore-snapshot", "EXACT_GIT_INPUT_UNAVAILABLE", err)
	}
	request := *action.WorkspaceSplitRequest
	environment := childEnvironment(gitDir, root)
	args := goooSplitPrepareArgs(root, request)
	prepared, err := runGoooSplitCommand(root, environment, args, &trace, pass, "executor")
	materialized.Executor = prepared.Observation
	if err != nil || prepared.Observation.ExitCode != 0 {
		return materialized, goooSplitProcessFailure("prepare", prepared, err)
	}
	proposal, err := readGoooSplitReceipt(prepared.Stdout)
	if err != nil {
		return materialized, goooSplitFailure("decode-proposal", "WORKSPACE_SPLIT_RECEIPT_INVALID", err)
	}
	if err := verifyGoooSplitProposal(root, request, action, proposal); err != nil {
		return materialized, goooSplitFailure("verify-proposal", "WORKSPACE_SPLIT_PROPOSAL_MISMATCH", err)
	}
	proposalPath := root + ".split-proposal.json"
	defer os.Remove(proposalPath)
	if err := os.WriteFile(proposalPath, prepared.Stdout, 0600); err != nil {
		return materialized, goooSplitFailure("save-proposal", "WORKSPACE_SPLIT_RECEIPT_UNAVAILABLE", err)
	}
	return applyAndVerifyGoooSplit(root, environment, proposalPath, request, proposal, plan, action, trace, pass, materialized)
}

func goooSplitPrepareArgs(root string, request workspacesplit.Request) []string {
	return []string{"package", "split-source", "--package", request.Package, "--source", request.Source,
		"--line-limit", strconv.Itoa(request.LineLimit), "--into", request.Into,
		"--discover-consumers", filepath.Join(root, filepath.FromSlash(request.DiscoveryRoot)),
		"--syntax-registry-root", root, "--json", filepath.Join(root, filepath.FromSlash(request.Workspace))}
}

func runGoooSplitCommand(root string, environment, args []string, trace *metaExecutionTrace, pass, kind string) (processResult, error) {
	actual := append([]string{"go", "run", "./cmd/gooo"}, args...)
	descriptor := make([]string, len(actual))
	for i, value := range actual {
		descriptor[i] = string(bytes.ReplaceAll([]byte(value), []byte(root), []byte("<workspace>")))
	}
	return runProcessObserved(root, environment, descriptor, actual, trace, pass, kind)
}

func goooSplitFailure(step, reason string, err error) *operationError {
	failure := newOperationError("WORKSPACE_SPLIT", step, reason, "KNOWN_CONTRADICTION", "preserve-workspace-split-counterexample")
	if err != nil {
		failure.diagnostics = []string{err.Error()}
	}
	return failure
}

func goooSplitProcessFailure(step string, process processResult, err error) *operationError {
	failure := newOperationError("WORKSPACE_SPLIT", step, "WORKSPACE_SPLIT_PROCESS_FAILED", "DIRECT_MISSING", "inspect-scoped-workspace-split")
	var observed struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(process.Stdout, &observed) == nil && observed.Error != "" {
		failure.diagnostics = []string{observed.Error}
	}
	if err != nil {
		failure.diagnostics = append(failure.diagnostics, err.Error())
	}
	return failure
}
