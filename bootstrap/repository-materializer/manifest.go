package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const manifestSchema = "gooo.repository-projection.v1"

func readManifest(physical string) (manifest, error) {
	name, err := resolvePath(physical, "projection/catalog/manifest.json")
	if err != nil {
		return manifest{}, err
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return manifest{}, err
	}
	model := manifest{}
	if err := json.Unmarshal(data, &model); err != nil {
		return manifest{}, err
	}
	if err := validateManifest(model); err != nil {
		return manifest{}, err
	}
	return model, nil
}

func validateManifest(model manifest) error {
	if model.Schema != manifestSchema || !validHex(model.SourceSHA, 40) {
		return fmt.Errorf("unsupported projection manifest identity")
	}
	previous := ""
	for _, item := range model.Entries {
		if item.Logical <= previous || item.Kind == "" || item.Backing == "" {
			return fmt.Errorf("unordered or incomplete manifest entry %q", item.Logical)
		}
		if _, err := resolvePath("/projection", item.Logical); err != nil {
			return err
		}
		if _, err := resolvePath("/projection", item.Backing); err != nil {
			return err
		}
		if !validHex(item.ObjectSHA, 64) || !validHex(item.ContentSHA, 64) {
			return fmt.Errorf("invalid digest for %s", item.Logical)
		}
		if item.Kind != "file" && item.Kind != "symlink" {
			return fmt.Errorf("unsupported projected kind %q", item.Kind)
		}
		previous = item.Logical
	}
	return nil
}

func validHex(value string, size int) bool {
	if len(value) != size {
		return false
	}
	for _, character := range strings.ToLower(value) {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
