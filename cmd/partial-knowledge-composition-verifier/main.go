package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	independent "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition/verify"
)

type options struct {
	repository string
	headSHA    string
	source     string
	cases      string
	receipt    string
	output     string
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
	flag.StringVar(&value.source, "source", "examples/partial-knowledge-composition/main.gooo", "Gooo source")
	flag.StringVar(&value.cases, "cases", "examples/partial-knowledge-composition/cases.json", "fixed case fixture")
	flag.StringVar(&value.receipt, "receipt", "", "producer receipt")
	flag.StringVar(&value.output, "output", "", "verification output")
	flag.Parse()
	return value
}

func run(value options) error {
	if value.headSHA == "" || value.receipt == "" || value.output == "" {
		return errors.New("-head-sha, -receipt, and -output are required")
	}
	source, err := os.ReadFile(value.source)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	fixture, err := os.ReadFile(value.cases)
	if err != nil {
		return fmt.Errorf("read cases: %w", err)
	}
	receipt, err := os.ReadFile(value.receipt)
	if err != nil {
		return fmt.Errorf("read receipt: %w", err)
	}
	report, err := independent.Verify(independent.Input{
		Repository: value.repository, HeadSHA: value.headSHA, SourcePath: value.source,
		Source: source, Fixture: fixture, Receipt: receipt,
	})
	if err != nil {
		return fmt.Errorf("independent verification: %w", err)
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode verification: %w", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(value.output, raw, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", value.output, err)
	}
	return nil
}
