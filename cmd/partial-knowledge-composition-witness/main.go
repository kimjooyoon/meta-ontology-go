package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	meta "github.com/kimjooyoon/meta-ontology-go/internal/meta/partialknowledgecomposition"
)

type options struct {
	repository string
	headSHA    string
	source     string
	cases      string
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
	flag.StringVar(&value.source, "source", meta.SourcePath, "Gooo source")
	flag.StringVar(&value.cases, "cases", "examples/partial-knowledge-composition/cases.json", "fixed case fixture")
	flag.StringVar(&value.output, "output", "", "receipt output")
	flag.Parse()
	return value
}

func run(value options) error {
	if value.headSHA == "" || value.output == "" {
		return errors.New("-head-sha and -output are required")
	}
	source, err := os.ReadFile(value.source)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	fixtureBytes, err := os.ReadFile(value.cases)
	if err != nil {
		return fmt.Errorf("read cases: %w", err)
	}
	var fixture meta.Fixture
	if err := json.Unmarshal(fixtureBytes, &fixture); err != nil {
		return fmt.Errorf("decode cases: %w", err)
	}
	receipt, err := meta.Evaluate(meta.Input{
		Repository: value.repository, HeadSHA: value.headSHA, SourcePath: value.source,
		Source: source, Fixture: fixture,
	})
	if err != nil {
		return fmt.Errorf("evaluate composition: %w", err)
	}
	if err := writeJSON(value.output, receipt); err != nil {
		return err
	}
	return nil
}

func writeJSON(filename string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(filename, raw, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	return nil
}
