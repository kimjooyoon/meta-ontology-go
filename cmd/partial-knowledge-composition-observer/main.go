package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/provider"
)

type options struct {
	repository, headSHA, sourceFile, sourcePath  string
	beforeTracked, beforeUntracked, beforeStatus string
	afterTracked, afterUntracked, afterStatus    string
	output                                       string
}

func main() {
	if err := run(parseOptions()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseOptions() options {
	value := options{}
	flag.StringVar(&value.repository, "repository", "kimjooyoon/meta-ontology-go", "repository identity")
	flag.StringVar(&value.headSHA, "head-sha", "", "exact checked-out head")
	flag.StringVar(&value.sourceFile, "source-file", "examples/partial-knowledge-composition/main.gooo", "physical Gooo source file")
	flag.StringVar(&value.sourcePath, "source-path", "examples/partial-knowledge-composition/main.gooo", "logical Gooo source path")
	flag.StringVar(&value.beforeTracked, "before-tracked", "", "pre-observation tracked snapshot")
	flag.StringVar(&value.beforeUntracked, "before-untracked", "", "pre-observation untracked snapshot")
	flag.StringVar(&value.beforeStatus, "before-status", "", "pre-observation status snapshot")
	flag.StringVar(&value.afterTracked, "after-tracked", "", "post-observation tracked snapshot")
	flag.StringVar(&value.afterUntracked, "after-untracked", "", "post-observation untracked snapshot")
	flag.StringVar(&value.afterStatus, "after-status", "", "post-observation status snapshot")
	flag.StringVar(&value.output, "output", "", "raw evidence output path")
	flag.Parse()
	return value
}

func run(value options) error {
	if value.headSHA == "" || value.output == "" || value.beforeTracked == "" || value.beforeUntracked == "" || value.beforeStatus == "" || value.afterTracked == "" || value.afterUntracked == "" || value.afterStatus == "" {
		return errors.New("-head-sha, all snapshot paths, and -output are required")
	}
	read := func(path string) ([]byte, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read snapshot %s: %w", path, err)
		}
		return data, nil
	}
	source, err := os.ReadFile(value.sourceFile)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	beforeTracked, err := read(value.beforeTracked)
	if err != nil {
		return err
	}
	beforeUntracked, err := read(value.beforeUntracked)
	if err != nil {
		return err
	}
	beforeStatus, err := read(value.beforeStatus)
	if err != nil {
		return err
	}
	afterTracked, err := read(value.afterTracked)
	if err != nil {
		return err
	}
	afterUntracked, err := read(value.afterUntracked)
	if err != nil {
		return err
	}
	afterStatus, err := read(value.afterStatus)
	if err != nil {
		return err
	}
	receipt, err := provider.Observe(provider.Input{
		Repository: value.repository, HeadSHA: value.headSHA, SourcePath: value.sourcePath, Source: source,
		BeforeTracked: beforeTracked, BeforeUntracked: beforeUntracked, BeforeStatus: beforeStatus,
		AfterTracked: afterTracked, AfterUntracked: afterUntracked, AfterStatus: afterStatus,
	})
	if err != nil {
		return fmt.Errorf("observe raw evidence: %w", err)
	}
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return fmt.Errorf("encode raw evidence: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(value.output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(value.output, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write raw evidence: %w", err)
	}
	return nil
}
