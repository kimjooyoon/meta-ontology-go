package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagedebugexperiment"
)

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("language-debug-experiment", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "input path")
	outputPath := flags.String("output", "", "output path")
	checkPath := flags.String("check", "", "expected report path")
	if flags.Parse(args) != nil || *inputPath == "" || (*outputPath == "" && *checkPath == "") {
		return 2
	}
	data, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	input, err := languagedebugexperiment.DecodeInput(data)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report, err := languagedebugexperiment.Evaluate(input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if *checkPath != "" && !matches(*checkPath, report) {
		return 1
	}
	if *outputPath != "" {
		encoded, _ := languagedebugexperiment.Marshal(report)
		if os.WriteFile(*outputPath, encoded, 0o644) != nil {
			return 2
		}
	}
	fmt.Fprintf(stdout, "language debug: %s %d/%d\n", report.Decision,
		report.Summary.Coordinates.Satisfied, report.Summary.Coordinates.Total)
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}

func matches(path string, report languagedebugexperiment.Report) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var expected languagedebugexperiment.Report
	return json.Unmarshal(data, &expected) == nil && expected.Digest == report.Digest
}
