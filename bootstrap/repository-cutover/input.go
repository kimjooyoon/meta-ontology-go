package main

import (
	"fmt"
	"path/filepath"
)

func resolveConfig(input cutoverConfig) (cutoverConfig, error) {
	if input.expectedSHA == "" {
		return input, fmt.Errorf("expected SHA is required")
	}
	paths := []struct {
		name  string
		value *string
	}{
		{"root", &input.root}, {"physical root", &input.physical},
		{"authority manifest", &input.authority}, {"evidence", &input.evidence},
	}
	if input.apply {
		paths = append(paths, struct {
			name  string
			value *string
		}{"backup", &input.backup})
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
	if input.root == input.physical || input.root == input.backup {
		return input, fmt.Errorf("cutover roots must be distinct")
	}
	return input, nil
}
