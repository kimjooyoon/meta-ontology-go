package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

const (
	packageResolveUsage       = "usage: gooo package <resolve|execute|replay|resume> ..."
	workspaceSchema           = "gooo/package-workspace-manifest/v1"
	workspaceMaxSourceCount   = 512
	workspaceMaxSourceBytes   = 1 << 20
	workspaceMaxSourceSetSize = 16 << 20
)

type workspaceManifest struct {
	Schema   string                   `json:"schema"`
	Entry    packageruntime.EntrySpec `json:"entry"`
	Packages []workspacePackage       `json:"packages"`
}

type workspacePackage struct {
	Path    string   `json:"path"`
	Name    string   `json:"name"`
	Imports []string `json:"imports"`
	Sources []string `json:"sources"`
}

type workspaceResolutionReceipt struct {
	Schema         string                 `json:"schema"`
	Decision       string                 `json:"decision"`
	Manifest       string                 `json:"manifest"`
	ManifestDigest string                 `json:"manifest_digest,omitempty"`
	Result         *packageruntime.Result `json:"result,omitempty"`
	Error          string                 `json:"error,omitempty"`
}

func runPackageCommand(args []string, reader SourceReader, stdout, stderr io.Writer) int {
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
	root := filepath.Dir(manifestPath)
	runtimeManifest := packageruntime.Manifest{
		Schema: packageruntime.ManifestSchema,
		Entry:  manifest.Entry,
	}
	sourceCount, sourceBytes := 0, 0
	for _, declared := range manifest.Packages {
		pkg := packageruntime.PackageSpec{
			Path: declared.Path, Name: declared.Name, Imports: append([]string(nil), declared.Imports...),
		}
		for _, sourcePath := range declared.Sources {
			relative, pathErr := workspaceSourcePath(sourcePath)
			if pathErr != nil {
				return packageResolutionFailure(jsonMode, manifestPath, pathErr, stderr, stdout)
			}
			filename := filepath.Join(root, relative)
			content, readErr := readSource(reader, filename)
			if readErr != nil {
				return packageResolutionFailure(jsonMode, manifestPath, fmt.Errorf("source %q: %w", sourcePath, readErr), stderr, stdout)
			}
			sourceCount++
			sourceBytes += len(content)
			if sourceCount > workspaceMaxSourceCount {
				return packageResolutionFailure(jsonMode, manifestPath, fmt.Errorf("workspace declares more than %d source files", workspaceMaxSourceCount), stderr, stdout)
			}
			if len(content) > workspaceMaxSourceBytes {
				return packageResolutionFailure(jsonMode, manifestPath, fmt.Errorf("source %q exceeds %d bytes", sourcePath, workspaceMaxSourceBytes), stderr, stdout)
			}
			if sourceBytes > workspaceMaxSourceSetSize {
				return packageResolutionFailure(jsonMode, manifestPath, fmt.Errorf("workspace sources exceed %d bytes total", workspaceMaxSourceSetSize), stderr, stdout)
			}
			pkg.Sources = append(pkg.Sources, packageruntime.Source{
				Filename: filepath.ToSlash(relative), Content: string(content),
			})
		}
		runtimeManifest.Packages = append(runtimeManifest.Packages, pkg)
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
	var manifest workspaceManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return workspaceManifest{}, fmt.Errorf("decode workspace manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return workspaceManifest{}, fmt.Errorf("workspace manifest contains trailing JSON")
		}
		return workspaceManifest{}, fmt.Errorf("decode workspace manifest trailing data: %w", err)
	}
	if manifest.Schema != workspaceSchema {
		return workspaceManifest{}, fmt.Errorf("workspace manifest schema must be %q", workspaceSchema)
	}
	if len(manifest.Packages) == 0 || len(manifest.Packages) > 64 {
		return workspaceManifest{}, fmt.Errorf("workspace manifest must declare 1 to 64 packages")
	}
	for _, pkg := range manifest.Packages {
		if len(pkg.Sources) == 0 || len(pkg.Sources) > workspaceMaxSourceCount {
			return workspaceManifest{}, fmt.Errorf("package %q must declare 1 to %d source files", pkg.Path, workspaceMaxSourceCount)
		}
	}
	return manifest, nil
}

func workspaceSourcePath(source string) (string, error) {
	if strings.TrimSpace(source) == "" || filepath.IsAbs(source) {
		return "", fmt.Errorf("workspace source path %q must be relative", source)
	}
	clean := filepath.Clean(filepath.FromSlash(source))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("workspace source path %q must stay inside the workspace", source)
	}
	if !strings.HasSuffix(clean, ".gooo") && !strings.HasSuffix(clean, ".gooo.fixture") {
		return "", fmt.Errorf("workspace source path %q must name a .gooo source or .gooo.fixture", source)
	}
	return clean, nil
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
