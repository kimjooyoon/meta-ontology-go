package causality

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DiscoverReport(roots []string, mode string) ([]byte, string, error) {
	type candidate struct {
		data []byte
		path string
	}
	candidates := make(map[string]candidate)
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		if _, err := os.Stat(root); err != nil {
			continue
		}
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if path != root && (entry.Name() == ".git" || entry.Name() == "node_modules") {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) != ".json" {
				return nil
			}
			info, err := entry.Info()
			if err != nil || info.Size() > 16<<20 {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, _, err := parseInputReport(data, mode); err != nil {
				return nil
			}
			digest := digestBytes(data)
			if _, exists := candidates[digest]; !exists {
				candidates[digest] = candidate{data: data, path: path}
			}
			return nil
		})
		if err != nil {
			return nil, "", fmt.Errorf("discover reports under %q: %w", root, err)
		}
	}
	if len(candidates) == 0 {
		return nil, "", fmt.Errorf("no %s %s report found", mode, InputReportSchema)
	}
	if len(candidates) != 1 {
		paths := make([]string, 0, len(candidates))
		for digest, candidate := range candidates {
			paths = append(paths, digest+":"+candidate.path)
		}
		sort.Strings(paths)
		return nil, "", fmt.Errorf("ambiguous %s reports: %s", mode, strings.Join(paths, ", "))
	}
	for _, candidate := range candidates {
		return candidate.data, candidate.path, nil
	}
	panic("unreachable")
}
