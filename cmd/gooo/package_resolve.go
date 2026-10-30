package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

const (
	packageResolveUsage       = "usage: gooo package <resolve|interface|inspect|origins|rename-activity|split-source|generate|execute|replay|resume|construct|session> ..."
	workspaceSchema           = workspaceinput.Schema
	workspaceMaxSourceCount   = workspaceinput.MaxSourceCount
	workspaceMaxSourceBytes   = workspaceinput.MaxSourceBytes
	workspaceMaxSourceSetSize = workspaceinput.MaxSourceSetBytes
)

type workspaceManifest = workspaceinput.Definition
type workspacePackage = workspaceinput.Package

type workspaceResolutionReceipt struct {
	Schema         string                 `json:"schema"`
	Decision       string                 `json:"decision"`
	Manifest       string                 `json:"manifest"`
	ManifestDigest string                 `json:"manifest_digest,omitempty"`
	Result         *packageruntime.Result `json:"result,omitempty"`
	Error          string                 `json:"error,omitempty"`
}

func runPackageCommand(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "generate" {
		return runPackageGenerate(args[1:], reader, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "split-source" {
		return runWorkspaceSourceSplit(args[1:], reader, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "origins" {
		return runPackageOrigins(args[1:], stdout, stderr)
	}
	if len(args) > 0 && args[0] == "rename-activity" {
		return runWorkspaceActivityRename(args[1:], reader, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "inspect" {
		return runPackageInspect(args[1:], reader, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "interface" {
		return runPackageInterface(args[1:], reader, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "construct" {
		return runPackageConstruct(args[1:], reader, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "resume" {
		return runPackageSaved(args[1:], reader, stdout, stderr, true)
	}
	if len(args) > 0 && args[0] == "replay" {
		return runPackageReplay(args[1:], reader, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "execute" {
		return runPackageExecute(args[1:], reader, stdout, stderr)
	}
	if len(args) == 0 || args[0] != "resolve" {
		return reportUsage(false, stdout, stderr, "package", packageResolveUsage)
	}
	args, jsonMode := parseJSONFlag(args[1:])
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return reportUsage(jsonMode, stdout, stderr, "package", packageResolveUsage)
	}
	manifestPath := args[0]
	manifestBytes, err := readSource(reader, manifestPath)
	if err != nil {
		return packageResolutionFailure(jsonMode, manifestPath, err, stderr, stdout)
	}
	if int64(len(manifestBytes)) > maxInputBytes {
		return packageResolutionFailure(jsonMode, manifestPath, inputLimitError(maxInputBytes), stderr, stdout)
	}
	manifest, err := decodeWorkspaceManifest(manifestBytes)
	if err != nil {
		return packageResolutionFailure(jsonMode, manifestPath, err, stderr, stdout)
	}
	runtimeManifest, err := loadPackageSources(reader, manifestPath, manifest)
	if err != nil {
		return packageResolutionFailure(jsonMode, manifestPath, err, stderr, stdout)
	}
	result, err := packageruntime.Run(runtimeManifest)
	if err != nil {
		return packageResolutionFailure(jsonMode, manifestPath, err, stderr, stdout)
	}
	if jsonMode {
		receipt := workspaceResolutionReceipt{
			Schema: "gooo/package-workspace-resolution-receipt/v1", Decision: "PASS",
			Manifest: filepath.ToSlash(manifestPath), ManifestDigest: workspaceDigest(manifestBytes), Result: &result,
		}
		if err := json.NewEncoder(stdout).Encode(receipt); err != nil {
			fmt.Fprintf(stderr, "gooo package resolve: write receipt: %v\n", err)
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "resolved Gooo package graph: packages=%d entry=%s.%s init_order=%s digest=%s\n",
		len(result.Image.Packages), result.Image.Entry.PackagePath, result.Image.Entry.Activity,
		strings.Join(result.Image.InitOrder, ","), result.ResultDigest)
	for _, pkg := range result.Image.Packages {
		fmt.Fprintf(stdout, "\npackage %s (%s)\n", pkg.Path, pkg.Name)
		if len(pkg.Imports) == 0 {
			fmt.Fprintln(stdout, "  imports: (none)")
		} else {
			fmt.Fprintf(stdout, "  imports: %s\n", strings.Join(pkg.Imports, ", "))
		}
		for _, export := range pkg.Exports {
			switch export.Kind {
			case "entity":
				fmt.Fprintf(stdout, "  entity %s id=%s\n", export.Name, export.ID)
			case "activity":
				fmt.Fprintf(stdout, "  activity %s(%s) -> %s\n", export.Name,
					strings.Join(export.InputTypes, ", "), export.OutputType)
			}
		}
		for _, binding := range pkg.Bindings {
			fmt.Fprintf(stdout, "  binding %s.%s.%s -> %s.%s.%s (%s)\n",
				binding.ProducerPackage, binding.ProducerActivity, binding.ProducerPort,
				binding.ConsumerPackage, binding.ConsumerActivity, binding.ConsumerPort, binding.EntityID)
		}
	}
	fmt.Fprintln(stdout, "\nNote: resolution checks package wiring; it does not execute activity bodies.")
	return exitOK
}

func decodeWorkspaceManifest(data []byte) (workspaceManifest, error) {
	return workspaceinput.Decode(data)
}

func workspaceSourcePath(source string) (string, error) {
	return workspaceinput.SourcePath(source)
}

func packageResolutionFailure(jsonMode bool, manifest string, cause error, stderr, stdout io.Writer) int {
	if jsonMode {
		receipt := workspaceResolutionReceipt{
			Schema: "gooo/package-workspace-resolution-receipt/v1", Decision: "FAIL_CLOSED",
			Manifest: filepath.ToSlash(manifest), Error: cause.Error(),
		}
		if err := json.NewEncoder(stdout).Encode(receipt); err != nil {
			fmt.Fprintf(stderr, "gooo package resolve: write failure receipt: %v\n", err)
		}
		return exitFailure
	}
	fmt.Fprintf(stderr, "gooo package resolve: %v\n", cause)
	return exitFailure
}

func workspaceDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
