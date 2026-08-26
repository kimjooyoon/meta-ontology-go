package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func resolveConfig(input activationConfig) (activationConfig, error) {
	if input.expectedSHA == "" {
		return input, fmt.Errorf("expected SHA is required")
	}
	paths := []struct {
		name  string
		value *string
	}{
		{"root", &input.root}, {"logical root", &input.logical},
		{"storage root", &input.storage}, {"Git directory", &input.gitDir},
		{"Git index", &input.gitIndex}, {"materialization evidence", &input.materialization},
		{"activation evidence", &input.evidence},
	}
	for _, item := range paths {
		if *item.value == "" {
			return input, fmt.Errorf("%s is required", item.name)
		}
		absolute, err := filepath.Abs(*item.value)
		if err != nil {
			return input, err
		}
		*item.value = filepath.Clean(absolute)
	}
	if input.root == input.logical || input.root == input.storage {
		return input, fmt.Errorf("activation roots must be distinct")
	}
	return input, nil
}

func readSourceEvidence(path, expectedSHA string) (sourceEvidence, error) {
	var proof sourceEvidence
	data, err := os.ReadFile(path)
	if err != nil {
		return proof, err
	}
	if err := json.Unmarshal(data, &proof); err != nil {
		return proof, err
	}
	if proof.Schema != "gooo.repository-materialization-evidence.v1" || proof.CurrentSHA != expectedSHA {
		return proof, fmt.Errorf("materialization identity is not exact")
	}
	if proof.Entries != proof.Restored {
		return proof, fmt.Errorf("materialization is incomplete")
	}
	for _, item := range proof.Indicators {
		if item.Blocking && item.Value > item.Limit {
			return proof, fmt.Errorf("materialization indicator %s blocks activation", item.ID)
		}
	}
	return proof, nil
}
