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
	density := flag.String("density-report", "", "exact density report")
	expected := flag.String("expected-sha", "", "required source SHA")
	output := flag.String("output", "", "extraction report output")
	flag.Parse()
	if err := run(*root, *plan, *density, *expected, *output); err != nil {
		log.Fatal(err)
	}
}

func run(root, plan, density, expected, output string) error {
	if root == "" || plan == "" || density == "" || expected == "" || output == "" {
		return fmt.Errorf("root, plan, density-report, expected-sha, and output are required")
	}
	plans, residual, err := loadExtractionInputs(plan, density, expected)
	if err != nil {
		return err
	}
	recipes, err := loadRecipes()
	if err != nil {
		return err
	}
	staged, subjects, unhandled, err := stageExtractions(root, plans, residual, recipes)
	if err != nil {
		return err
	}
	report := extractionEvidence(expected, subjects, unhandled)
	if err := writeExtractionReport(filepath.Clean(output), report); err != nil {
		return err
	}
	if err := requireHandled(report); err != nil {
		return err
	}
	if err := commitStaged(staged); err != nil {
		return err
	}
	fmt.Printf("function-extractor: residual=%d applied=%d\n", len(residual), len(subjects))
	return nil
}
