package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func writeStore(work string, model manifest,
	objects map[string]*storedObject, files []trackedFile) (string, error) {
	root := filepath.Join(work, "physical")
	names := make([]string, 0, len(objects))
	for name := range objects {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		object := objects[name]
		target := filepath.Join(root, filepath.FromSlash(object.backing))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, object.data, 0o644); err != nil {
			return "", err
		}
	}
	for _, file := range files {
		if file.backing == "" {
			continue
		}
		if file.kind != "file" {
			return "", fmt.Errorf("retained symlink is unsupported: %s", file.logical)
		}
		target := filepath.Join(root, filepath.FromSlash(file.backing))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, file.data, os.FileMode(file.mode)); err != nil {
			return "", err
		}
	}
	catalog := filepath.Join(root, "projection", "catalog")
	if err := os.MkdirAll(catalog, 0o755); err != nil {
		return "", err
	}
	encoded, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", err
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(filepath.Join(catalog, "manifest.json"), encoded, 0o644); err != nil {
		return "", err
	}
	return root, nil
}
