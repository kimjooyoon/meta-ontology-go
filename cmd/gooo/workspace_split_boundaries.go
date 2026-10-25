package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func inspectWorkspaceSourceSplit(options workspaceSourceSplitOptions, reader SourceReader) (workspaceSourceSplitReceipt, error) {
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
	observed, err := packageruntime.InspectSourceSplit(loaded, options.packagePath, filepath.ToSlash(source))
	if err != nil {
		return receipt, err
	}
	receipt.Decision, receipt.Boundaries = "OBSERVED", &observed
	return receipt, nil
}

func printSourceSplitBoundaries(out io.Writer, observed packageruntime.SourceSplitBoundaries) error {
	if _, err := fmt.Fprintf(out, "%s: %d physical lines, %d declaration cuts; consumer admission unassessed\n",
		observed.InputFilename, observed.Input.Lines, len(observed.Candidates)); err != nil {
		return err
	}
	for _, candidate := range observed.Candidates {
		if _, err := fmt.Fprintf(out, "before %s: %d + %d physical lines (id=%s)\n",
			candidate.Name, candidate.Left.Lines, candidate.Right.Lines, candidate.BeforeID); err != nil {
			return err
		}
	}
	return nil
}
