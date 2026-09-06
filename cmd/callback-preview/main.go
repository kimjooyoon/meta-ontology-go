package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func main() {
	root := flag.String("root", "", "read-only repository root")
	logical := flag.String("logical", "cmd/language-readiness-witness/predecessor-selection/pagination_test.go", "logical pagination fixture path")
	output := flag.String("output", "", "caller-owned preview JSON output")
	flag.Parse()
	if *root == "" || *output == "" {
		fatal("root and output are required")
	}
	preview, err := extractor.PreviewBoundedPaginationCallback(*root, filepath.ToSlash(*logical))
	if err != nil {
		fatal(err.Error())
	}
	payload, err := json.MarshalIndent(preview, "", "  ")
	if err != nil {
		fatal(err.Error())
	}
	if err := writePreviewOutput(*root, *output, append(payload, '\n')); err != nil {
		fatal(err.Error())
	}
}

func writePreviewOutput(root, output string, payload []byte) error {
	path, err := safePreviewOutputPath(root, output)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create preview output directory: %w", err)
	}
	path, err = safePreviewOutputPath(root, path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write preview output: %w", err)
	}
	return nil
}

func safePreviewOutputPath(root, output string) (string, error) {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	rootPath, err = filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", fmt.Errorf("resolve repository root symlinks: %w", err)
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return "", fmt.Errorf("resolve preview output: %w", err)
	}
	if withinPreviewPath(rootPath, outputPath) {
		return "", fmt.Errorf("preview output must be outside the repository root")
	}
	if info, err := os.Lstat(outputPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("preview output must not be a symlink")
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect preview output: %w", err)
	}
	if info, err := os.Stat(outputPath); err == nil {
		if aliases, aliasErr := previewOutputAliasesRepositoryFile(rootPath, info); aliasErr != nil {
			return "", aliasErr
		} else if aliases {
			return "", fmt.Errorf("preview output aliases a repository input")
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat preview output: %w", err)
	}
	canonical, err := canonicalPreviewOutputPath(outputPath)
	if err != nil {
		return "", err
	}
	if withinPreviewPath(rootPath, canonical) {
		return "", fmt.Errorf("preview output resolves inside the repository root")
	}
	return outputPath, nil
}

func previewOutputAliasesRepositoryFile(root string, outputInfo os.FileInfo) (bool, error) {
	var aliases bool
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if aliases || info.IsDir() {
			return nil
		}
		if os.SameFile(outputInfo, info) {
			aliases = true
		}
		return nil
	})
	return aliases, err
}

func canonicalPreviewOutputPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve preview output path: %w", err)
	}
	missing := []string{}
	for current := path; ; current = filepath.Dir(current) {
		info, statErr := os.Lstat(current)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return "", fmt.Errorf("preview output path traverses a symlink")
			}
			resolved, evalErr := filepath.EvalSymlinks(current)
			if evalErr != nil {
				return "", fmt.Errorf("resolve preview output parent: %w", evalErr)
			}
			for _, component := range slices.Backward(missing) {
				resolved = filepath.Join(resolved, component)
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(statErr) {
			return "", fmt.Errorf("inspect preview output parent: %w", statErr)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("preview output has no existing parent")
		}
		missing = append(missing, filepath.Base(current))
	}
}

func withinPreviewPath(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
