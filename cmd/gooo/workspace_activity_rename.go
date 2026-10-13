package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const workspaceActivityRenameUsage = "usage: gooo package rename-activity --package <path> --id <activity-id> --to <name> [--json] <workspace.json>"

type workspaceActivityRenameOptions struct {
	packagePath, activityID, to, manifestPath string
	jsonMode                                  bool
}

type workspaceActivityRenameReceipt struct {
	Schema           string                                      `json:"schema"`
	Decision         string                                      `json:"decision"`
	ManifestPath     string                                      `json:"manifest_path"`
	ManifestDigest   string                                      `json:"manifest_digest,omitempty"`
	UpdatedWorkspace *workspaceManifest                          `json:"updated_workspace,omitempty"`
	Result           *workspaceexecution.WorkspaceActivityRename `json:"result,omitempty"`
	Error            string                                      `json:"error,omitempty"`
}

func runWorkspaceActivityRename(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseWorkspaceActivityRename(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, workspaceActivityRenameUsage)
		return exitUsage
	}
	receipt, err := prepareWorkspaceActivityRename(options, reader)
	if writeErr := writeWorkspaceActivityRename(receipt, options.jsonMode, stdout, stderr); writeErr != nil {
		fmt.Fprintln(stderr, "gooo package rename-activity:", writeErr)
		return exitFailure
	}
	if err != nil {
		return exitFailure
	}
	return exitOK
}

func parseWorkspaceActivityRename(args []string, stderr io.Writer) (workspaceActivityRenameOptions, error) {
	flags := flag.NewFlagSet("package rename-activity", flag.ContinueOnError)
	flags.SetOutput(stderr)
	options := workspaceActivityRenameOptions{}
	flags.StringVar(&options.packagePath, "package", "", "package containing the target activity")
	flags.StringVar(&options.activityID, "id", "", "stable semantic ID of the activity")
	flags.StringVar(&options.to, "to", "", "new source activity name")
	flags.BoolVar(&options.jsonMode, "json", false, "print the complete source proposal")
	if err := flags.Parse(args); err != nil {
		return options, err
	}
	if strings.TrimSpace(options.packagePath) == "" || strings.TrimSpace(options.activityID) == "" ||
		strings.TrimSpace(options.to) == "" || flags.NArg() != 1 {
		return options, fmt.Errorf("package, ID, new name and one workspace manifest are required")
	}
	options.manifestPath = flags.Arg(0)
	return options, nil
}

func prepareWorkspaceActivityRename(options workspaceActivityRenameOptions, reader SourceReader) (workspaceActivityRenameReceipt, error) {
	receipt := workspaceActivityRenameReceipt{Schema: "gooo/workspace-activity-rename-receipt/v1",
		Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath}
	manifest, digest, err := loadWorkspaceRenameManifest(reader, options.manifestPath)
	receipt.ManifestDigest = digest
	if err != nil {
		receipt.Error = err.Error()
		return receipt, err
	}
	sources, err := loadPackageSources(reader, options.manifestPath, manifest)
	if err != nil {
		receipt.Error = err.Error()
		return receipt, err
	}
	result, err := workspaceexecution.RenameWorkspaceActivity(context.Background(), sources,
		options.packagePath, options.activityID, options.to)
	if err != nil {
		receipt.Error = err.Error()
		return receipt, err
	}
	manifest.Entry = result.Manifest.Entry
	receipt.Decision, receipt.Result, receipt.UpdatedWorkspace = "PROPOSED", &result, &manifest
	return receipt, nil
}
