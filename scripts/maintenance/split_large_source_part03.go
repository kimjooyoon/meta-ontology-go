package main

import (
	"os"
	"path/filepath"
	"sort"
)

func collectTargets(opts options) ([]string, error) {
	if len(opts.targets) > 0 {
		paths := make([]string, 0, len(opts.targets))
		for _, target := range opts.targets {
			resolved := filepath.Clean(filepath.Join(opts.root, target))
			if stat, err := os.Stat(resolved); err == nil && stat.IsDir() {
				if err := filepath.Walk(resolved, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if info.Name() == ".git" || info.Name() == "vendor" {
							return filepath.SkipDir
						}
						return nil
					}
					if isSourceFile(info.Name()) {
						paths = append(paths, path)
					}
					return nil
				}); err != nil {
					return nil, err
				}
				continue
			}
			if isSourceFile(resolved) {
				paths = append(paths, resolved)
			}
		}
		sort.Strings(paths)
		return dedupe(paths), nil
	}
	paths := make([]string, 0)
	if err := filepath.Walk(opts.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if isSourceFile(info.Name()) {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}
func isSourceFile(path string) bool {
	return filepath.Ext(path) == ".go" || filepath.Ext(path) == ".gooo"
}
