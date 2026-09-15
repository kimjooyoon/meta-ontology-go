package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func readCatalog(settings cutoverConfig) (catalog, []string, string, string, int, error) {
	var model catalog
	candidatePath := filepath.Join(settings.physical, "projection", "catalog", "manifest.json")
	candidate, err := os.ReadFile(candidatePath)
	if err != nil {
		return model, nil, "", "", 1, err
	}
	authority, err := os.ReadFile(settings.authority)
	if err != nil {
		return model, nil, "", "", 1, err
	}
	if err := json.Unmarshal(candidate, &model); err != nil {
		return model, nil, "", "", 1, err
	}
	if model.Schema != "gooo.repository-projection.v1" {
		return model, nil, "", "", 1, fmt.Errorf("unsupported catalog schema")
	}
	expected := map[string]bool{"projection/catalog/manifest.json": true}
	for _, item := range model.Entries {
		if !safePath(item.Backing) {
			return model, nil, "", "", 1, fmt.Errorf("unsafe backing path %q", item.Backing)
		}
		expected[item.Backing] = true
	}
	paths := make([]string, 0, len(expected))
	for name := range expected {
		paths = append(paths, name)
	}
	sort.Strings(paths)
	drift := 0
	if !bytes.Equal(candidate, authority) {
		drift = 1
	}
	return model, paths, digest(authority), digest(candidate), drift, nil
}

func safePath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return path != "" && path == clean && path != "." && !filepath.IsAbs(path) && !strings.HasPrefix(path, "../")
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
