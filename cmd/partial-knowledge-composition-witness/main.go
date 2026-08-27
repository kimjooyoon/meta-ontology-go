package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	meta "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition"
)

type options struct {
	repository, headSHA, sourceFile, sourcePath, evidence, output, intervention string
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
	flag.StringVar(&value.sourceFile, "source-file", meta.SourcePath, "physical Gooo source file")
	flag.StringVar(&value.sourcePath, "source-path", meta.SourcePath, "logical Gooo source path")
	flag.StringVar(&value.evidence, "evidence", "", "CI raw evidence receipt")
	flag.StringVar(&value.output, "output", "", "receipt output path")
	flag.StringVar(&value.intervention, "intervention", string(meta.InterventionNone), "none, semantic, or comment-only")
	flag.Parse()
	return value
}

func run(value options) error {
	if value.headSHA == "" || value.evidence == "" || value.output == "" {
		return errors.New("-head-sha, -evidence, and -output are required")
	}
	source, err := os.ReadFile(value.sourceFile)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	rawEvidence, err := os.ReadFile(value.evidence)
	if err != nil {
		return fmt.Errorf("read raw evidence: %w", err)
	}
	receipt, err := meta.Evaluate(meta.Input{Repository: value.repository, HeadSHA: value.headSHA, SourcePath: value.sourcePath, Source: source, RawEvidence: rawEvidence, Intervention: meta.InterventionMode(value.intervention)})
	if err != nil {
		return fmt.Errorf("produce partial-knowledge receipt: %w", err)
	}
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return fmt.Errorf("encode receipt: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(value.output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(value.output, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}
	return nil
}
