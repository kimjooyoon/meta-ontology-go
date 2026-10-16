package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func preparePaths(cfg *config) error {
	if cfg.Root == "" {
		return fmt.Errorf("counterexamples require an explicit subject repository")
	}
	root, err := filepath.EvalSymlinks(cfg.Root)
	if err != nil {
		return err
	}
	cfg.Root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	for _, path := range []*string{&cfg.Output, &cfg.Scratch} {
		if *path == "" {
			return fmt.Errorf("counterexample outputs and scratch require explicit external paths")
		}
		absolute, err := filepath.Abs(*path)
		if err != nil {
			return err
		}
		parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
		if err != nil {
			return err
		}
		canonical := filepath.Join(parent, filepath.Base(absolute))
		relative, err := filepath.Rel(cfg.Root, canonical)
		if err != nil || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("counterexample output must remain outside the repository")
		}
		if _, err := os.Lstat(canonical); !os.IsNotExist(err) {
			return fmt.Errorf("counterexample output must be a new owned directory")
		}
		*path = canonical
	}
	if cfg.Output == cfg.Scratch {
		return fmt.Errorf("counterexample output and scratch must be distinct")
	}
	return os.Mkdir(cfg.Output, 0750)
}
