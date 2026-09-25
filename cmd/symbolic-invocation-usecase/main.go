package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/symbolicinvocationusecase"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("symbolic-invocation-usecase", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "input evidence path")
	outputPath := flags.String("output", "", "output report path")
	checkPath := flags.String("check", "", "expected report path")
	if flags.Parse(args) != nil || *inputPath == "" || (*outputPath == "" && *checkPath == "") {
		return 2
	}
	data, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	input, err := symbolicinvocationusecase.DecodeInput(data)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report, err := symbolicinvocationusecase.Evaluate(input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if *checkPath != "" && !matches(*checkPath, report) {
		return 1
	}
	if *outputPath != "" {
		encoded, marshalErr := symbolicinvocationusecase.Marshal(report)
		if marshalErr != nil || os.WriteFile(*outputPath, encoded, 0o644) != nil {
			return 2
		}
	}
	fmt.Fprintf(stdout, "symbolic invocation use case: %s %d/%d\n", report.Decision,
		report.Summary.Coordinates.Satisfied, report.Summary.Coordinates.Total)
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}

func matches(path string, report symbolicinvocationusecase.Report) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var expected symbolicinvocationusecase.Report
	return json.Unmarshal(data, &expected) == nil && expected.Digest == report.Digest
}
