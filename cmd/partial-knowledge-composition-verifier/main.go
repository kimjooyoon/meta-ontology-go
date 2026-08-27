package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/verify"
)

type options struct {
	repository, headSHA, sourceFile, sourcePath, evidence, receipt, output, intervention string
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
	flag.StringVar(&value.evidence, "evidence", "", "CI raw evidence receipt")
	flag.StringVar(&value.receipt, "receipt", "", "producer receipt")
	flag.StringVar(&value.output, "output", "", "verification output")
	flag.StringVar(&value.intervention, "intervention", "none", "none, semantic, or comment-only")
	flag.Parse()
	return value
}

func run(value options) error {
	if value.headSHA == "" || value.evidence == "" || value.receipt == "" || value.output == "" {
		return errors.New("-head-sha, -evidence, -receipt, and -output are required")
	}
	source, err := os.ReadFile(value.sourceFile)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	receipt, err := os.ReadFile(value.receipt)
	if err != nil {
		return fmt.Errorf("read receipt: %w", err)
	}
	rawEvidence, err := os.ReadFile(value.evidence)
	if err != nil {
		return fmt.Errorf("read raw evidence: %w", err)
	}
	report, err := independent.Verify(independent.Input{Repository: value.repository, HeadSHA: value.headSHA, SourcePath: value.sourcePath, Source: source, RawEvidence: rawEvidence, InterventionMode: value.intervention, Receipt: receipt})
	if err != nil {
		return fmt.Errorf("independent verification: %w", err)
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode verification: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(value.output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(value.output, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write verification: %w", err)
	}
	return nil
}
