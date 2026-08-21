package main

import (
	"fmt"
	"path/filepath"
)

func execute(input config) error {
	settings, err := resolveConfig(input)
	if err != nil {
		return err
	}
	head, err := exactHead(settings.root, settings.expectedSHA)
	if err != nil {
		return err
	}
	model, err := readManifest(settings.physical)
	if err != nil {
		return err
	}
	if err := prepareWork(settings); err != nil {
		return err
	}
	destination := filepath.Join(settings.work, "materialized")
	restored, err := restoreEntries(model, settings.physical, destination)
	if err != nil {
		return err
	}
	state, err := buildLogicalIndex(settings, model, destination, head)
	if err != nil {
		return err
	}
	if err := writeEvidence(settings, model, head, restored, state); err != nil {
		return err
	}
	fmt.Printf("repository-materializer: restored=%d tree=%s replacement=%t\n", restored, state.TreeOID, state.Replacement != "")
	return nil
}

func unexpectedPhysical(tracked map[string]string, model manifest) int {
	expected := map[string]bool{"projection/catalog/manifest.json": true}
	for _, item := range model.Entries {
		expected[item.Backing] = true
	}
	unexpected := 0
	for name := range tracked {
		if !expected[name] {
			unexpected++
		}
	}
	return unexpected
}
