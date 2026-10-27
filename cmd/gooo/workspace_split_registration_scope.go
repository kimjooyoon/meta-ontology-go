package main

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle/embeduse"
)

func splitRegistrationTarget(root, name string) (string, error) {
	if !fs.ValidPath(name) || strings.ContainsAny(name, "\\\x00:") {
		return "", fmt.Errorf("syntax registration requires canonical relative paths")
	}
	target := filepath.Join(root, filepath.FromSlash(name))
	parent, err := filepath.EvalSymlinks(filepath.Dir(target))
	if err != nil {
		return "", err
	}
	if parent != root && !pathContained(root, parent) {
		return "", fmt.Errorf("syntax registration target escapes repository")
	}
	if parent != filepath.Dir(target) {
		return "", fmt.Errorf("syntax registration needs physical directory paths before resolving embedded consumers")
	}
	return filepath.Join(parent, filepath.Base(target)), nil
}

// Embed patterns cannot ascend outside their declaring directory. Inspect all
// ancestor packages in the explicit repository scope, including test consumers.
// Dynamic reads and consumers outside this scope are not discovered here.
func splitRegistrationEmbedInputs(root string, candidate syntaxregistration.PartitionRegistrationCandidate) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	directories, sources := map[string]bool{}, map[string][]byte{}
	for name, raw := range candidate.Outputs {
		if !strings.HasSuffix(name, ".gooo") {
			continue
		}
		if _, err := splitRegistrationTarget(root, name); err != nil {
			return nil, err
		}
		sources[name] = raw
		for directory := filepath.Dir(filepath.Join(root, filepath.FromSlash(name))); ; directory = filepath.Dir(directory) {
			directories[directory] = true
			if directory == root {
				break
			}
		}
	}
	inputs := map[string][]byte{}
	for _, directory := range slices.Sorted(maps.Keys(directories)) {
		embedded, err := captureSplitAncestorGo(root, directory, inputs)
		if err != nil {
			return nil, err
		}
		if embedded {
			if err := requireUnembeddedSplit(ctx, root, directory, sources); err != nil {
				return nil, err
			}
		}
	}
	return inputs, ctx.Err()
}

func captureSplitAncestorGo(root, directory string, inputs map[string][]byte) (bool, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return false, err
	}
	embedded := false
	total := 0
	for _, raw := range inputs {
		total += len(raw)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		name := filepath.Join(directory, entry.Name())
		if len(inputs) >= 1024 {
			return false, fmt.Errorf("source ancestor scope exceeds 1024 Go files")
		}
		raw, err := readRegularFile(name, int64(min(sourcebundle.MaxSourceBytes, sourcebundle.MaxTotalBytes-total)))
		if err != nil {
			return false, err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return false, err
		}
		inputs[filepath.ToSlash(relative)] = raw
		total += len(raw)
		file, err := parser.ParseFile(token.NewFileSet(), name, raw, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return false, err
		}
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:embed ") || strings.HasPrefix(comment.Text, "//go:embed\t") {
					embedded = true
				}
			}
		}
	}
	return embedded, nil
}

func requireUnembeddedSplit(ctx context.Context, root, directory string, sources map[string][]byte) error {
	var projected []sourcebundle.Source
	for _, name := range slices.Sorted(maps.Keys(sources)) {
		relative, err := filepath.Rel(directory, filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		if fs.ValidPath(filepath.ToSlash(relative)) {
			projected = append(projected, sourcebundle.Source{Filename: filepath.ToSlash(relative), Content: string(sources[name])})
		}
	}
	for _, projection := range [][]sourcebundle.Source{nil, projected} {
		report, err := embeduse.Inspect(ctx, directory, projection)
		if err != nil {
			return fmt.Errorf("inspect embedded source consumers: %w", err)
		}
		if len(report.IgnoredGoFiles) != 0 {
			return fmt.Errorf("embedded source scope includes inactive Go files; native consumer admission is required")
		}
		for _, consumer := range report.Consumers {
			for _, index := range consumer.Members {
				name, err := filepath.Rel(root, filepath.Join(directory, filepath.FromSlash(report.Files[index].Filename)))
				if err != nil {
					return err
				}
				if _, affected := sources[filepath.ToSlash(name)]; affected {
					return fmt.Errorf("embedded consumer %s in %s requires native admission before joint publication", consumer.Name, directory)
				}
			}
		}
	}
	return nil
}
