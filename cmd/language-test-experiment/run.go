package main

import (
	"flag"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagetestexperiment"
)

func run(args []string) int {
	flags := flag.NewFlagSet("language-test-experiment", flag.ContinueOnError)
	inputPath := flags.String("input", "", "input evidence JSON")
	outputPath := flags.String("output", "", "write evaluated report")
	checkPath := flags.String("check", "", "compare evaluated report")
	if err := flags.Parse(args); err != nil || *inputPath == "" || (*outputPath == "" && *checkPath == "") {
		return 2
	}
	input, err := languagetestexperiment.ReadInput(*inputPath)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	report, err := languagetestexperiment.Evaluate(input)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if *outputPath != "" {
		if err := languagetestexperiment.WriteReport(*outputPath, report); err != nil {
			fmt.Println(err)
			return 1
		}
	}
	if *checkPath != "" {
		if err := languagetestexperiment.CheckReport(*checkPath, report); err != nil {
			fmt.Println(err)
			return 1
		}
	}
	fmt.Printf("language test: %s %d/%d\n", report.Decision, report.Summary.Coordinates.Satisfied, report.Summary.Coordinates.Total)
	if report.Decision != "PASS" {
		return 1
	}
	return 0
}
