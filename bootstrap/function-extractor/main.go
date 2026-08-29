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
	report.BackupCleanup = backupCleanupObservation{Status: "PENDING", Attempted: transactionBackupCount(transaction.files)}
	provisional, err := createProvisionalReportPath(filepath.Clean(output), report)
	if err != nil {
		return rollbackReportTransaction(transaction, provisional, err)
	}
	report.BackupCleanup = removeTransactionBackups(transaction.files)
	if err := removeReport(provisional); err != nil {
		return rollbackReportTransaction(transaction, provisional, err)
	}
	if err := writeExtractionReport(filepath.Clean(output), report); err != nil {
		return rollbackReportTransaction(transaction, provisional, err)
	}
	if report.BackupCleanup.Status != "PASS" {
		return fmt.Errorf("backup cleanup incomplete: %d/%d removed", report.BackupCleanup.Removed, report.BackupCleanup.Attempted)
	}
	fmt.Printf("function-extractor: residual=%d applied=%d created=%d\n",
		len(residual), len(subjects), createdCount(subjects))
	return nil
}

func rollbackReportTransaction(transaction stagedTransaction, provisional string, cause error) error {
	cleanupErr := removeReport(provisional)
	rollbackErr := rollbackTransactions(transaction.files, len(transaction.files))
	if cleanupErr != nil {
		if rollbackErr != nil {
			return fmt.Errorf("%w; provisional cleanup failed: %v; rollback failed: %v", cause, cleanupErr, rollbackErr)
		}
		return fmt.Errorf("%w; provisional cleanup failed: %v", cause, cleanupErr)
	}
	if rollbackErr != nil {
		return fmt.Errorf("%w; rollback failed: %v", cause, rollbackErr)
	}
	return cause
}
