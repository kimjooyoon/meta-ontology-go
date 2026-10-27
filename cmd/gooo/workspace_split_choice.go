package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func parseSplitLineLimit(options *workspaceSourceSplitOptions, value string) error {
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 || options.lineLimit != 0 {
		return fmt.Errorf("supply one positive --line-limit")
	}
	options.lineLimit = limit
	return nil
}

func prepareChosenWorkspaceSplit(options workspaceSourceSplitOptions, reader SourceReader) (workspaceSourceSplitReceipt, error) {
	receipt := workspaceSourceSplitReceipt{Schema: "gooo/workspace-source-split-receipt/v1",
		Decision: "FAIL_CLOSED", ManifestPath: options.manifestPath}
	definition, digest, err := loadWorkspaceRenameManifest(reader, options.manifestPath)
	receipt.ManifestDigest = digest
	if err != nil {
		return receipt, err
	}
	loaded, err := workspaceinput.Load(reader, options.manifestPath, definition)
	if err != nil {
		return receipt, err
	}
	source, err := workspaceinput.SourcePath(options.source)
	if err != nil {
		return receipt, err
	}
	choice, err := packageruntime.ChooseSourceSplit(loaded, options.packagePath, filepath.ToSlash(source), options.lineLimit)
	if err != nil {
		return receipt, err
	}
	receipt.Selection = &choice
	if choice.State == "ALREADY_WITHIN_LIMIT" {
		receipt.Decision = "UNCHANGED"
		return receipt, nil
	}
	if choice.State != "SELECTED" || choice.Boundary == nil {
		return receipt, fmt.Errorf("source split selection: %s", choice.State)
	}
	options.lineLimit, options.before = 0, choice.Boundary.BeforeID
	proposal, err := prepareWorkspaceSourceSplit(options, reader)
	if err != nil {
		return receipt, err
	}
	if proposal.ManifestDigest != digest || proposal.Result == nil ||
		proposal.Result.InputSHA256 != choice.InputSHA256 || proposal.Result.InputImageDigest != choice.InputImageDigest {
		return receipt, fmt.Errorf("workspace changed between boundary selection and preparation")
	}
	proposal.Selection = &choice
	return proposal, nil
}
