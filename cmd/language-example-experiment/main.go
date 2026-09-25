package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageexampleexperiment"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	options, err := parseOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	input := languageexampleexperiment.Input{ExpectedHead: options.expectedHead}
	loads := []struct {
		path   string
		target any
	}{
		{options.contract, &input.Contract}, {options.golden, &input.Golden},
		{options.artifact, &input.Artifact}, {options.replay, &input.Replay},
		{options.unknownEmitter, &input.UnknownEmitter}, {options.profile, &input.Profile},
	}
	for _, load := range loads {
		if err := readJSON(load.path, load.target); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	report := languageexampleexperiment.Evaluate(input)
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(options.out, append(payload, '\n'), 0o640); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("language example: %s %d/%d\n", report.Decision,
		report.Summary.Coordinates.Satisfied, report.Summary.Coordinates.Total)
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}

func readJSON(path string, target any) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

var _ artifactemit.Artifact
