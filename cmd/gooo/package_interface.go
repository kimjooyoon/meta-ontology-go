package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type workspaceInterfaceReceipt struct {
	Schema         string                    `json:"schema"`
	Decision       string                    `json:"decision"`
	Manifest       string                    `json:"manifest"`
	ManifestDigest string                    `json:"manifest_digest,omitempty"`
	Interface      *packageruntime.Interface `json:"interface,omitempty"`
	Error          string                    `json:"error,omitempty"`
}

func runPackageInterface(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
		return reportUsage(jsonMode, stdout, stderr, "package interface", "usage: gooo package interface [--json] <workspace.json>")
	}
	receipt := workspaceInterfaceReceipt{Schema: "gooo/package-workspace-interface-receipt/v1",
		Decision: "FAIL_CLOSED", Manifest: filepath.ToSlash(args[0])}
	view, manifestDigest, err := describeWorkspace(reader, args[0])
	receipt.ManifestDigest = manifestDigest
	if err == nil {
		receipt.Decision, receipt.Interface = "PASS", &view
	} else {
		receipt.Error = err.Error()
	}
	if jsonMode {
		if writeErr := json.NewEncoder(stdout).Encode(receipt); writeErr != nil {
			return exitFailure
		}
	} else if err != nil {
		fmt.Fprintf(stderr, "gooo package interface: %v\n", err)
	} else if writeErr := writePackageInterface(stdout, view); writeErr != nil {
		return exitFailure
	}
	if err != nil {
		return exitFailure
	}
	return exitOK
}

func describeWorkspace(reader SourceReader, path string) (packageruntime.Interface, string, error) {
	raw, err := readSource(reader, path)
	if err != nil {
		return packageruntime.Interface{}, "", err
	}
	if int64(len(raw)) > maxInputBytes {
		return packageruntime.Interface{}, "", inputLimitError(maxInputBytes)
	}
	digest := workspaceDigest(raw)
	manifest, err := decodeWorkspaceManifest(raw)
	if err != nil {
		return packageruntime.Interface{}, digest, err
	}
	sources, err := loadPackageSources(reader, path, manifest)
	if err != nil {
		return packageruntime.Interface{}, digest, err
	}
	view, err := packageruntime.Describe(sources)
	return view, digest, err
}

func writePackageInterface(output io.Writer, view packageruntime.Interface) error {
	var text strings.Builder
	for _, pkg := range view.Packages {
		fmt.Fprintf(&text, "package %s (%s)\n", pkg.Path, pkg.Name)
		for _, declaration := range pkg.Declarations {
			if declaration.Kind == "activity" {
				fmt.Fprintf(&text, "  activity %s(%s) -> %s id=%s\n", declaration.Name,
					strings.Join(declaration.Inputs, ", "), declaration.Output, declaration.ID)
				continue
			}
			fmt.Fprintf(&text, "  %s %s id=%s\n", declaration.Shape, declaration.Name, declaration.ID)
			for _, field := range declaration.Fields {
				fmt.Fprintf(&text, "    %s: %s %s %s id=%s\n", field.Name, field.TypeID, field.Presence, field.Cardinality, field.ID)
			}
		}
	}
	_, err := io.WriteString(output, text.String())
	return err
}
