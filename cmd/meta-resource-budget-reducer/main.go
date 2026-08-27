package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageresourcebudget"
)

func main() {
	inputPath := flag.String("input", "", "producer observation input")
	outputPath := flag.String("output", "", "reduced report output")
	caseName := flag.String("case", "normal", "normal, over-budget, or missing-sample")
	checkPath := flag.String("check", "", "report to validate instead of writing")
	flag.Parse()
	if *inputPath == "" || (*outputPath == "" && *checkPath == "") || (*outputPath != "" && *checkPath != "") {
		fail("usage: meta-resource-budget-reducer -input INPUT (-output OUTPUT | -check REPORT) [-case NAME]")
	}
	input, err := languageresourcebudget.ReadInput(*inputPath)
	if err != nil {
		fail(err.Error())
	}
	if *checkPath != "" {
		report, err := languageresourcebudget.ReadReport(*checkPath)
		if err != nil {
			fail(err.Error())
		}
		if err := languageresourcebudget.ValidateReport(input, report); err != nil {
			fail(err.Error())
		}
		return
	}
	report := languageresourcebudget.Evaluate(input, *caseName)
	if err := languageresourcebudget.WriteReport(*outputPath, report); err != nil {
		fail(err.Error())
	}
	fmt.Printf("resource budget: case=%s decision=%s resolution=%s reason=%s samples=%d/%d\n", report.Case, report.Decision, report.Resolution, report.Reason, report.Summary.Samples, report.Summary.Operations)
	if report.Decision != "PASS" {
		os.Exit(1)
	}
}

func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(2) }
