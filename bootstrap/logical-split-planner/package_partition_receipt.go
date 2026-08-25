package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func applyPackageTextEdits(root string, recipe packagePartitionRecipe, writes map[string]bool) error {
	for _, rewrite := range recipe.Rewrites {
		name, err := sourcePath(root, rewrite.Path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		data, err = replaceExact(data, rewrite.Old, rewrite.New, rewrite.ExpectedCount)
		if err != nil {
			return fmt.Errorf("%s: %w", rewrite.Path, err)
		}
		if err := os.WriteFile(name, data, 0o644); err != nil {
			return err
		}
		writes[rewrite.Path] = true
	}
	for _, edit := range recipe.Ranges {
		name, err := sourcePath(root, edit.Path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		text, start := string(data), strings.Index(string(data), edit.Start)
		if start < 0 || strings.Count(text, edit.Start) != 1 {
			return fmt.Errorf("range start is not exact: %s", edit.Path)
		}
		end := len(text)
		if edit.End != "" {
			end = strings.Index(text[start+len(edit.Start):], edit.End)
			if end < 0 || strings.Count(text, edit.End) != 1 {
				return fmt.Errorf("range end is not exact: %s", edit.Path)
			}
			end += start + len(edit.Start)
		}
		text = text[:start] + edit.Replacement + text[end:]
		if err := os.WriteFile(name, []byte(text), 0o644); err != nil {
			return err
		}
		writes[edit.Path] = true
	}
	return nil
}

func requirePackageShape(root string, recipe packagePartitionRecipe) error {
	branch, err := sourcePath(root, recipe.Subject)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(branch)
	if err != nil || len(entries) != recipe.ExpectedShape.BranchEntries {
		return fmt.Errorf("partition branch shape mismatch")
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			return fmt.Errorf("partition branch mixes entry kinds")
		}
	}
	for leaf, expected := range recipe.ExpectedShape.Leaves {
		entries, err := os.ReadDir(branch + string(os.PathSeparator) + leaf)
		if err != nil || len(entries) != expected || len(entries) > recipe.ExpectedShape.MaxEntries {
			return fmt.Errorf("partition leaf shape mismatch: %s", leaf)
		}
	}
	return nil
}

func writePackagePartitionReceipt(name, sha string, recipe packagePartitionRecipe, writes map[string]bool) error {
	paths := make([]string, 0, len(writes))
	for path := range writes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	receipt := map[string]any{"schema": "gooo.go-package-partition-receipt.v1", "decision": "PASS", "resolution": "EXACT", "reason": "GO_PACKAGE_PARTITION_APPLIED", "source_sha": sha, "subject": recipe.Subject, "meta_operation": "partition-go-package", "moved_files": len(recipe.Moves), "created_files": len(recipe.Creates), "rewrites": len(recipe.Rewrites) + len(recipe.Ranges), "write_set": paths, "expected_shape": recipe.ExpectedShape, "effects": map[string]any{"repository_writes": 0, "mutation_authority": false, "disposable_writes": len(paths)}}
	sealed, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(sealed)
	receipt["digest"] = "sha256:" + hex.EncodeToString(digest[:])
	payload, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(payload, '\n'), 0o644)
}
