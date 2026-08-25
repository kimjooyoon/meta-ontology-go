package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func applyPackageRecipe(root string, recipe packagePartitionRecipe, writes map[string]bool) error {
	for _, move := range recipe.Moves {
		source, err := sourcePath(root, move.Source)
		if err != nil {
			return err
		}
		target, err := sourcePath(root, move.Destination)
		if err != nil {
			return err
		}
		info, err := os.Lstat(source)
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("move source is not a regular file: %s", move.Source)
		}
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			return fmt.Errorf("move destination already exists: %s", move.Destination)
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		if move.Package != "" {
			data, err = replaceExact(data, "package main", "package "+move.Package, 1)
			if err != nil {
				return fmt.Errorf("%s: %w", move.Source, err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, info.Mode().Perm()); err != nil {
			return err
		}
		if err := os.Remove(source); err != nil {
			return err
		}
		writes[move.Source], writes[move.Destination] = true, true
	}
	for _, create := range recipe.Creates {
		name, err := sourcePath(root, create.Path)
		if err != nil {
			return err
		}
		if _, err := os.Lstat(name); !os.IsNotExist(err) {
			return fmt.Errorf("create destination already exists: %s", create.Path)
		}
		if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(name, []byte(create.Content), 0o644); err != nil {
			return err
		}
		writes[create.Path] = true
	}
	return applyPackageTextEdits(root, recipe, writes)
}

func replaceExact(data []byte, old, replacement string, count int) ([]byte, error) {
	if old == "" || strings.Count(string(data), old) != count {
		return nil, fmt.Errorf("replacement count does not equal %d", count)
	}
	return []byte(strings.ReplaceAll(string(data), old, replacement)), nil
}
