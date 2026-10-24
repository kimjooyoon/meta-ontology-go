package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

const workspaceSourceSplitUsage = "usage: gooo package split-source --package <path> --source <file> --before <declaration-id> --into <file.gooo> [--json] <workspace.json>"

type workspaceSourceSplitOptions struct {
	packagePath, source, before, into, manifestPath string
	jsonMode                                        bool
}

type workspaceSourceSplitReceipt struct {
	Schema           string                      `json:"schema"`
	Decision         string                      `json:"decision"`
	ManifestPath     string                      `json:"manifest_path"`
	ManifestDigest   string                      `json:"manifest_digest,omitempty"`
	UpdatedWorkspace *workspaceManifest          `json:"updated_workspace,omitempty"`
	Result           *packageruntime.SourceSplit `json:"result,omitempty"`
	Error            string                      `json:"error,omitempty"`
}

func runWorkspaceSourceSplit(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseWorkspaceSourceSplit(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, workspaceSourceSplitUsage)
		return exitUsage
	}
	receipt, err := prepareWorkspaceSourceSplit(options, reader)
	if err != nil {
		receipt.Error = err.Error()
	}
	if options.jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if writeErr := encoder.Encode(receipt); writeErr != nil {
			fmt.Fprintln(stderr, "gooo package split-source:", writeErr)
			return exitFailure
		}
	} else if err != nil {
		fmt.Fprintln(stderr, "gooo package split-source:", err)
	} else {
		if _, err := fmt.Fprintf(stdout, "proposed %s: %s + %s before id=%s\nUse --json for sources, byte origins and the updated workspace.\n",
			options.packagePath, options.source, options.into, options.before); err != nil {
			return exitFailure
		}
	}
	if err != nil {
		return exitFailure
	}
	return exitOK
}

func parseWorkspaceSourceSplit(args []string, stderr io.Writer) (workspaceSourceSplitOptions, error) {
	flags := flag.NewFlagSet("package split-source", flag.ContinueOnError)
	flags.SetOutput(stderr)
	options := workspaceSourceSplitOptions{}
	flags.StringVar(&options.packagePath, "package", "", "package owning the source")
	flags.StringVar(&options.source, "source", "", "source path relative to the workspace")
	flags.StringVar(&options.before, "before", "", "stable ID of the first declaration to move")
	flags.StringVar(&options.into, "into", "", "new source path relative to the workspace")
	flags.BoolVar(&options.jsonMode, "json", false, "print the complete source proposal")
	if err := flags.Parse(args); err != nil {
		return options, err
	}
	for _, value := range []string{options.packagePath, options.source, options.before, options.into} {
		if strings.TrimSpace(value) == "" {
			return options, fmt.Errorf("package, source, boundary ID and destination are required")
		}
	}
	if flags.NArg() != 1 {
		return options, fmt.Errorf("one workspace manifest is required")
	}
	options.manifestPath = flags.Arg(0)
	return options, nil
}

func prepareWorkspaceSourceSplit(options workspaceSourceSplitOptions, reader SourceReader) (workspaceSourceSplitReceipt, error) {
	receipt := workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1",
		Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath}
	manifest, digest, err := loadWorkspaceRenameManifest(reader, options.manifestPath)
	receipt.ManifestDigest = digest
	if err != nil {
		return receipt, err
	}
	sources, err := loadPackageSources(reader, options.manifestPath, manifest)
	if err != nil {
		return receipt, err
	}
	result, err := packageruntime.SplitSource(sources, options.packagePath, options.source, options.before, options.into)
	if err != nil {
		return receipt, err
	}
	if err := checkSplitWorkspaceLimits(result.Manifest); err != nil {
		return receipt, err
	}
	for i := range manifest.Packages {
		if manifest.Packages[i].Path == options.packagePath {
			manifest.Packages[i].Sources = append(manifest.Packages[i].Sources, options.into)
		}
	}
	receipt.Decision, receipt.Result, receipt.UpdatedWorkspace = "PROPOSED", &result, &manifest
	return receipt, nil
}

func checkSplitWorkspaceLimits(manifest packageruntime.Manifest) error {
	count, size := 0, 0
	for _, pkg := range manifest.Packages {
		for _, source := range pkg.Sources {
			count++
			size += len(source.Content)
			if count > workspaceMaxSourceCount || len(source.Content) > workspaceMaxSourceBytes || size > workspaceMaxSourceSetSize {
				return fmt.Errorf("split output exceeds the workspace source count or byte limit")
			}
		}
	}
	return nil
}
