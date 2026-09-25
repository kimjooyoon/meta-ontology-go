package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
)

func main() {
	root := flag.String("root", "", "restored logical repository")
	plan := flag.String("plan", "", "exact logical split plan")
	expected := flag.String("expected-sha", "", "required source SHA")
	output := flag.String("output", "", "density report output")
	flag.Parse()
	if err := run(*root, *plan, *expected, *output); err != nil {
		log.Fatal(err)
	}
}

func run(root, plan, expected, output string) error {
	if root == "" || plan == "" || expected == "" || output == "" {
		return fmt.Errorf("root, plan, expected-sha, and output are required")
	}
	inputs, err := loadDensitySubjects(plan, expected)
	if err != nil {
		return err
	}
	results, err := rewriteSubjects(root, inputs)
	if err != nil {
		return err
	}
	report := densityReport(expected, results)
	if err := writeDensityReport(filepath.Clean(output), report); err != nil {
		return err
	}
	fmt.Printf("line-density-rewriter: subjects=%d applied=%d\n", len(results), appliedCount(results))
	return requireDensityClosure(report)
}
