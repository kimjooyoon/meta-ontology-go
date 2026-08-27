package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	expansion "github.com/kimjooyoon/meta-ontology-go/internal/meta/capabilityscopedexpansion"
)

type options struct {
	output, subject, pinnedFile, logicalInput, sandbox, repository string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	options, err := parseOptions()
	if err != nil {
		return err
	}
	provider, err := expansion.CaptureProvider(expansion.ProviderRequest{SubjectSHA: options.subject, PinnedFile: options.pinnedFile, LogicalInput: options.logicalInput, SandboxRoot: options.sandbox, RepositoryRoot: options.repository})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(options.output), 0o755); err != nil {
		return err
	}
	return os.WriteFile(options.output, append(provider, '\n'), 0o644)
}

func parseOptions() (options, error) {
	result := options{}
	flags := flag.NewFlagSet("capability-scoped-expansion-provider", flag.ContinueOnError)
	flags.StringVar(&result.output, "output", "", "raw provider observation artifact")
	flags.StringVar(&result.subject, "subject-sha", "", "exact CI subject SHA")
	flags.StringVar(&result.pinnedFile, "pinned-file", "", "CI-created file observed by the provider")
	flags.StringVar(&result.logicalInput, "logical-input", "", "CI-created deterministic logical input")
	flags.StringVar(&result.sandbox, "sandbox", "", "temporary sandbox")
	flags.StringVar(&result.repository, "repository-root", "", "repository root for tracked and untracked snapshots")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return result, err
	}
	if result.output == "" || result.subject == "" || result.pinnedFile == "" || result.logicalInput == "" || result.sandbox == "" || result.repository == "" {
		return result, fmt.Errorf("output, subject-sha, pinned-file, logical-input, sandbox, and repository-root are required")
	}
	return result, nil
}
