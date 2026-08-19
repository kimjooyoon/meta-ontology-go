package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func equalSlotBodies(previous, current map[string][]byte) bool {
	if len(previous) != len(current) {
		return false
	}
	for id, body := range previous {
		if string(body) != string(current[id]) {
			return false
		}
	}
	return true
}
func resolveManifestPath(root, raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("manifest path is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("manifest absolute path: %w", err)
	}
	manifestRoot, err := canonicalOutputRoot(filepath.Dir(abs))
	if err != nil {
		return "", err
	}
	return resolveOutputPath(manifestRoot, filepath.Base(abs))
}
