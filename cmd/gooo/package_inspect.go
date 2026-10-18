package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageInspectUsage = "usage: gooo package inspect [--json | --inputs-template] <workspace.json>"

type workspaceInspectionReceipt struct {
	Schema            string                            `json:"schema"`
	Decision          string                            `json:"decision"`
	Manifest          string                            `json:"manifest"`
	ManifestDigest    string                            `json:"manifest_digest,omitempty"`
	Program           *workspaceexecution.Program       `json:"program,omitempty"`
	ProgramExecutions int                               `json:"program_executions"`
	ModelPredictions  int                               `json:"model_predictions"`
	Capacity          bodyexecution.CompositionCapacity `json:"transport_capacity"`
	Error             string                            `json:"error,omitempty"`
	Scope             string                            `json:"scope"`
}

func runPackageInspect(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	inputTemplate := false
	clean := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--inputs-template" {
			if inputTemplate {
				return reportUsage(jsonMode, stdout, stderr, "package inspect", packageInspectUsage)
			}
			inputTemplate = true
			continue
		}
		clean = append(clean, arg)
	}
	args = clean
	if (jsonMode && inputTemplate) || len(args) != 1 || strings.TrimSpace(args[0]) == "" ||
		strings.HasPrefix(args[0], "-") {
		return reportUsage(jsonMode, stdout, stderr, "package inspect", packageInspectUsage)
	}
	receipt := workspaceInspectionReceipt{Schema: "gooo/workspace-source-inspection/v1", Decision: "FAIL_CLOSED",
		Capacity: bodyexecution.CompositionTransportCapacity(),
		Manifest: filepath.ToSlash(args[0]),
		Scope:    "source package graph and prepared entry call closure; no body construction, predicate evaluation, model prediction or native execution"}
	program, manifestDigest, err := inspectWorkspaceSource(reader, args[0])
	if inputTemplate {
		if err == nil {
			var raw []byte
			raw, err = workspaceexecution.InputTemplate(context.Background(), program)
			if err == nil {
				_, err = fmt.Fprintln(stdout, string(raw))
			}
		}
		if err != nil {
			fmt.Fprintf(stderr, "gooo package inspect: %v\n", err)
			return exitFailure
		}
		return exitOK
	}
	receipt.ManifestDigest = manifestDigest
	if err == nil {
		receipt.Decision, receipt.Program = "PREPARED", &program
	} else {
		receipt.Error = err.Error()
	}
	if jsonMode {
		if writeErr := json.NewEncoder(stdout).Encode(receipt); writeErr != nil {
			fmt.Fprintf(stderr, "gooo package inspect: %v\n", writeErr)
			return exitFailure
		}
	} else if err != nil {
		fmt.Fprintf(stderr, "gooo package inspect: %v\n", err)
	} else {
		calls := 0
		if program.PureCalls != nil {
			calls = len(program.PureCalls.Sites)
		}
		if _, writeErr := fmt.Fprintf(stdout, "prepared workspace entry: %s.%s packages=%d source_calls=%d\n",
			program.Entry.PackagePath, program.Entry.Activity, len(program.Workspace.Packages), calls); writeErr != nil {
			return exitFailure
		}
	}
	if err != nil {
		return exitFailure
	}
	return exitOK
}

func inspectWorkspaceSource(reader SourceReader, path string) (workspaceexecution.Program, string, error) {
	raw, err := readSource(reader, path)
	if err != nil {
		return workspaceexecution.Program{}, "", err
	}
	if int64(len(raw)) > maxInputBytes {
		return workspaceexecution.Program{}, "", inputLimitError(maxInputBytes)
	}
	digest := workspaceDigest(raw)
	manifest, err := decodeWorkspaceManifest(raw)
	if err != nil {
		return workspaceexecution.Program{}, digest, err
	}
	sources, err := loadPackageSources(reader, path, manifest)
	if err != nil {
		return workspaceexecution.Program{}, digest, err
	}
	program, err := workspaceexecution.Prepare(sources)
	return program, digest, err
}
