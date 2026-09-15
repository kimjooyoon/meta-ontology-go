package main

import (
	"fmt"
	"os"
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
