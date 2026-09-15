package main

import (
	"io/fs"
	"path/filepath"
	"sort"
)

func leafPaths(root string) ([]string, error) {
	paths := []string{}
	err := filepath.WalkDir(root, func(path string, item fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root || item.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func mismatch(left, right []string) int {
	set := map[string]int{}
	for _, name := range left {
		set[name]++
	}
	for _, name := range right {
		set[name]--
	}
	difference := 0
	for _, count := range set {
		if count < 0 {
			count = -count
		}
		difference += count
	}
	return difference
}

func unionCount(left, right []string) int {
	set := map[string]bool{}
	for _, names := range [][]string{left, right} {
		for _, name := range names {
			set[name] = true
		}
	}
	return len(set)
}
