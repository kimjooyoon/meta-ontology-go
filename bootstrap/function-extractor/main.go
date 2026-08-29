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
	fixedPoint := flag.Bool("fixed-point", false, "include generic extraction residuals from the current plan")
	flag.Parse()
	if err := run(*root, *plan, *density, *expected, *output, *fixedPoint); err != nil {
		log.Fatal(err)
	}
}

func run(root, plan, density, expected, output string, fixedPoint bool) error {
	if root == "" || plan == "" || density == "" || expected == "" || output == "" {
		return fmt.Errorf("root, plan, density-report, expected-sha, and output are required")
	}
	plans, residual, err := loadExtractionInputs(plan, density, expected, fixedPoint)
	if err != nil {
		return err
	}
	recipes, err := loadRecipes()
	if err != nil {
		return err
	}
	staged, subjects, unhandled, failures, err := stageExtractions(root, plans, residual, recipes)
	if err != nil {
		return err
	}
	report := extractionEvidence(expected, subjects, unhandled, failures)
	if err := requireHandled(report); err != nil {
		return err
	}
	transaction, err := commitStaged(staged)
	if err != nil {
		return err
	}
	report.NamespaceReplacements = transaction.receipts
	if err := writeExtractionReport(filepath.Clean(output), report); err != nil {
		rollbackTransactions(transaction.files, len(transaction.files))
		return err
	}
	if err := removeTransactionBackups(transaction.files); err != nil {
		return err
	}
	fmt.Printf("function-extractor: residual=%d applied=%d created=%d\n",
		len(residual), len(subjects), createdCount(subjects))
	return nil
}
