package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

const workspaceSourceSplitUsage = "usage: gooo package split-source (--package <path> --source <file> --before <declaration-id> --into <file.gooo> [--consumer-workspace <workspace.json>] | --apply <proposal.json>) [--json] <workspace.json>"

type workspaceSourceSplitOptions struct {
	packagePath, source, before, into, manifestPath string
	applyReceipt                                    string
	consumerWorkspaces                              []string
	jsonMode                                        bool
}

type workspaceSourceSplitReceipt struct {
	Schema           string                           `json:"schema"`
	Decision         string                           `json:"decision"`
	ManifestPath     string                           `json:"manifest_path"`
	ManifestDigest   string                           `json:"manifest_digest,omitempty"`
	UpdatedWorkspace *workspaceManifest               `json:"updated_workspace,omitempty"`
	Result           *packageruntime.SourceSplit      `json:"result,omitempty"`
	Error            string                           `json:"error,omitempty"`
	Application      *workspaceSourceSplitApplication `json:"application,omitempty"`
	Consumers        []workspaceSourceSplitConsumer   `json:"consumers,omitempty"`
}

func runWorkspaceSourceSplit(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseWorkspaceSourceSplit(args, stderr)
	if err != nil {
		fmt.Fprintln(stderr, workspaceSourceSplitUsage)
		return exitUsage
	}
	var receipt workspaceSourceSplitReceipt
	if options.applyReceipt != "" {
		receipt, err = applyWorkspaceSourceSplit(options, defaultAtomicFileOps())
	} else {
		receipt, err = prepareWorkspaceSourceSplit(options, reader)
	}
	if err != nil {
		receipt.Error = err.Error()
	}
	if options.jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if writeErr := encoder.Encode(sourceSplitPublicReceipt(receipt)); writeErr != nil {
			fmt.Fprintln(stderr, "gooo package split-source:", writeErr)
			return exitFailure
		}
	} else if err != nil {
		fmt.Fprintln(stderr, "gooo package split-source:", err)
	} else {
		r := receipt.Result
		if _, err := fmt.Fprintf(stdout, "%s %s: %s + %s before id=%s\nUse --json for sources, byte origins and the updated workspace.\n",
			strings.ToLower(receipt.Decision), r.PackagePath, r.InputFilename, r.Files[1].Filename, r.BeforeID); err != nil {
			return exitFailure
		}
	}
	if err != nil {
		return exitFailure
	}
	return exitOK
}

// The two changed files and external source list are the wire proposal. Keep
// the runtime manifest available to Go callers without duplicating every
// unchanged source in a saved CLI receipt. Legacy full manifests remain readable.
func sourceSplitPublicReceipt(receipt workspaceSourceSplitReceipt) workspaceSourceSplitReceipt {
	if receipt.Result != nil {
		result := *receipt.Result
		result.Manifest = packageruntime.Manifest{}
		receipt.Result = &result
	}
	receipt.Consumers = append([]workspaceSourceSplitConsumer(nil), receipt.Consumers...)
	for i := range receipt.Consumers {
		if receipt.Consumers[i].Result != nil {
			result := *receipt.Consumers[i].Result
			result.Manifest = packageruntime.Manifest{}
			receipt.Consumers[i].Result = &result
		}
	}
	return receipt
}

func parseWorkspaceSourceSplit(args []string, stderr io.Writer) (workspaceSourceSplitOptions, error) {
	flags := flag.NewFlagSet("package split-source", flag.ContinueOnError)
	flags.SetOutput(stderr)
	options := workspaceSourceSplitOptions{}
	flags.StringVar(&options.packagePath, "package", "", "package owning the source")
	flags.StringVar(&options.source, "source", "", "source path relative to the workspace")
	flags.StringVar(&options.before, "before", "", "stable ID of the first declaration to move")
	flags.StringVar(&options.into, "into", "", "new source path relative to the workspace")
	flags.StringVar(&options.applyReceipt, "apply", "", "apply a saved proposal after checking current inputs")
	flags.Func("consumer-workspace", "additional workspace reading this source (repeatable)", func(value string) error {
		if strings.TrimSpace(value) == "" || len(options.consumerWorkspaces) >= maxSplitConsumerWorkspaces {
			return fmt.Errorf("supply at most %d nonempty additional workspace paths", maxSplitConsumerWorkspaces)
		}
		options.consumerWorkspaces = append(options.consumerWorkspaces, value)
		return nil
	})
	flags.BoolVar(&options.jsonMode, "json", false, "print the complete source proposal")
	if err := flags.Parse(args); err != nil {
		return options, err
	}
	if flags.NArg() != 1 {
		return options, fmt.Errorf("one workspace manifest is required")
	}
	options.manifestPath = flags.Arg(0)
	if options.applyReceipt != "" && len(options.consumerWorkspaces) != 0 {
		return options, fmt.Errorf("--apply reads consumer workspaces from the saved proposal")
	}
	for _, value := range []string{options.packagePath, options.source, options.before, options.into} {
		if options.applyReceipt != "" {
			if value != "" {
				return options, fmt.Errorf("--apply cannot be combined with proposal options")
			}
			continue
		}
		if strings.TrimSpace(value) == "" {
			return options, fmt.Errorf("package, source, boundary ID and destination are required")
		}
	}
	return options, nil
}

func prepareWorkspaceSourceSplit(options workspaceSourceSplitOptions, reader SourceReader) (workspaceSourceSplitReceipt, error) {
	receipt := workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1",
		Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath}
	if len(options.consumerWorkspaces) != 0 {
		return prepareSharedWorkspaceSourceSplit(options, reader)
	}
	manifest, digest, err := loadWorkspaceRenameManifest(reader, options.manifestPath)
	receipt.ManifestDigest = digest
	if err != nil {
		return receipt, err
	}
	proposal, err := workspaceinput.PrepareSplit(reader, options.manifestPath, manifest, workspaceinput.SplitRequest{
		PackagePath: options.packagePath, Source: options.source, BeforeID: options.before, Into: options.into})
	if err != nil {
		return receipt, err
	}
	receipt.Decision, receipt.Result, receipt.UpdatedWorkspace = "PROPOSED", &proposal.Result, &proposal.Definition
	return receipt, nil
}
