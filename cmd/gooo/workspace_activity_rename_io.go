package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func loadWorkspaceRenameManifest(reader SourceReader, path string) (workspaceManifest, string, error) {
	raw, err := readSource(reader, path)
	if err != nil {
		return workspaceManifest{}, "", err
	}
	digest := workspaceDigest(raw)
	manifest, err := decodeWorkspaceManifest(raw)
	return manifest, digest, err
}

func writeWorkspaceActivityRename(receipt workspaceActivityRenameReceipt, jsonMode bool, stdout, stderr io.Writer) error {
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(receipt)
	}
	if receipt.Error != "" {
		_, err := fmt.Fprintln(stderr, "gooo package rename-activity:", receipt.Error)
		return err
	}
	r := receipt.Result
	if _, err := fmt.Fprintf(stdout, "proposed %s.%s -> %s id=%s references=%d/%d files=%d entry_updated=%t\n",
		r.PackagePath, r.From, r.To, r.ActivityID, r.UpdatedReferences, r.ResolvedReferences, len(r.Files), r.EntryUpdated); err != nil {
		return err
	}
	for _, file := range r.Files {
		if _, err := fmt.Fprintf(stdout, "  %s: %s edits=%d\n", file.PackagePath, file.Filename, len(file.Edits)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(stdout, "Use --json to save the source proposal and updated workspace manifest.")
	return err
}
