package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

type goooSplitRuntime struct {
	CasesSHA256     string                           `json:"cases_sha256"`
	GeneratedSHA256 string                           `json:"generated_sha256"`
	GoToolSHA256    string                           `json:"go_tool_sha256"`
	FiniteTotal     int                              `json:"finite_total"`
	FinitePassed    int                              `json:"finite_passed"`
	Traces          []bodyexecution.CompositionTrace `json:"traces"`
}

func verifyGoooSplitRuntime(root string, environment []string, request workspacesplit.Request,
	trace metaExecutionTrace, pass string) (goooSplitRuntime, processResult, *operationError) {
	casesPath := filepath.Join(root, filepath.FromSlash(request.Cases))
	file, err := os.Open(casesPath)
	if err != nil {
		return goooSplitRuntime{}, processResult{}, goooSplitUnavailable("read-cases", "WORKSPACE_CASES_UNAVAILABLE", err)
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, bodyexecution.CompositionInputMaxBytes+1))
	if err != nil {
		return goooSplitRuntime{}, processResult{}, goooSplitUnavailable("read-cases", "WORKSPACE_CASES_UNAVAILABLE", err)
	}
	cases, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		return goooSplitRuntime{}, processResult{}, goooSplitFailure("read-cases", "WORKSPACE_CASES_INVALID", err)
	}
	args := []string{"package", "execute", "--cases", casesPath, "--json", filepath.Join(root, filepath.FromSlash(request.Workspace))}
	process, runErr := runGoooSplitCommand(root, environment, args, &trace, pass, "verifier")
	if runErr != nil || process.Observation.ExitCode != 0 {
		return goooSplitRuntime{}, process, goooSplitProcessFailure("execute-cases", process, runErr)
	}
	observed, err := decodeGoooSplitRuntime(process.Stdout, digestBytes(raw), len(cases.Cases))
	if err != nil {
		return observed, process, goooSplitFailure("verify-cases", "WORKSPACE_RUNTIME_ORACLE_MISMATCH", err)
	}
	return observed, process, nil
}

func decodeGoooSplitRuntime(raw []byte, casesDigest string, total int) (goooSplitRuntime, error) {
	var receipt struct {
		Decision    string                     `json:"decision"`
		Error       string                     `json:"error"`
		CasesDigest string                     `json:"cases_digest"`
		Result      *workspaceexecution.Result `json:"result"`
	}
	if err := completeness.CheckJSON(raw, 32<<20); err != nil {
		return goooSplitRuntime{}, err
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return goooSplitRuntime{}, err
	}
	if receipt.Result == nil || receipt.Error != "" || receipt.Decision != "PASS" || receipt.CasesDigest != casesDigest {
		return goooSplitRuntime{}, fmt.Errorf("runtime receipt is incomplete or uses different cases")
	}
	runtime := receipt.Result.Runtime
	if total < 1 || runtime.FiniteTotal != total || runtime.FinitePassed != total ||
		!runtime.ProjectionReplayed || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 || runtime.Failure != "" || len(runtime.Traces) != total ||
		!validBareDigest(strings.TrimPrefix(runtime.GeneratedSHA256, "sha256:")) ||
		!validBareDigest(strings.TrimPrefix(runtime.GoToolSHA256, "sha256:")) {
		return goooSplitRuntime{}, fmt.Errorf("independent finite expectations or deterministic replay failed")
	}
	for index, trace := range runtime.Traces {
		if trace.CaseIndex != index {
			return goooSplitRuntime{}, fmt.Errorf("finite runtime case population differs")
		}
	}
	return goooSplitRuntime{CasesSHA256: casesDigest, GeneratedSHA256: runtime.GeneratedSHA256,
		GoToolSHA256: runtime.GoToolSHA256, FiniteTotal: total, FinitePassed: runtime.FinitePassed, Traces: runtime.Traces}, nil
}
